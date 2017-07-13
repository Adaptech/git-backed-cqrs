package main

import (
	"net/http"
	"os"
	"os/exec"
	"html/template"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"crypto/rand"
	"path"
	"reflect"
	"strings"
	"log"
)

type TodoList struct {
	Name string
}
type TodoListCreated struct {
	Message
	Name string
}
type TodoListItemCreated struct {
	Message
	Name string
}
type Message struct {
	Id string
}
type TodoListCreate struct {
	Message
	Name string
}
func listTodoListsPageHandler(w http.ResponseWriter, r *http.Request) {
	todoLists := []TodoList{}
	file, err := ioutil.ReadFile("storage/projections/todoLists")
	if err != nil { json.Unmarshal(file, &todoLists) }
	PageTemplates.ExecuteTemplate(w, "todoLists.html", todoLists)

}
var PageTemplates = template.Must(template.ParseGlob("templates/*.html"))
var Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
func pseudo_uuid() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}
func main() {

	makeStorage()
	http.HandleFunc("/Todolists", listTodoListsPageHandler)
	http.HandleFunc("/createTodoList", createTodoList)
	http.HandleFunc("/createTodoListForm", showCreateTodoListForm)
	http.ListenAndServe(":8080", nil)
}
func showCreateTodoListForm(writer http.ResponseWriter, request *http.Request) {
	PageTemplates.ExecuteTemplate(writer, "createTodoList.html", nil)
}
func createTodoList(writer http.ResponseWriter, request *http.Request) {
	createTodoCommand := &TodoListCreate{
		Message {
			pseudo_uuid(),
		},
		request.FormValue("name"),
	}
	handleCreateTodoList(createTodoCommand)
}
func handleCreateTodoList(todoListCreate *TodoListCreate) {
	fmt.Printf("handleCreateTodoList Id = %v Name = %v\n", todoListCreate.Id, todoListCreate.Name)
	todoListCreated := &TodoListCreated{
		Message {
			todoListCreate.Id,
		},
		todoListCreate.Name,
	}
	storeEvent(todoListCreated)
}
func storeEvent(todoListCreated *TodoListCreated) {
	streamDir := path.Join("storage", todoListCreated.Id)
	Info.Println(fmt.Printf("storeEvent streamDir = %v\n", streamDir))
	dirs, _ := ioutil.ReadDir(streamDir)
	if len(dirs) == 0 {
		os.Mkdir(streamDir, 0755)
	}
	Info.Println(fmt.Printf("storeEvent dirs = %v len = %v\n", dirs, len(dirs)))
	seqNum := len(dirs) + 1
	cmdStrs := strings.Split(fmt.Sprintf("%s", reflect.TypeOf(todoListCreated)), ".")
	eventString := cmdStrs[len(cmdStrs) - 1]
	fileName := path.Join(streamDir, fmt.Sprintf("%06d", seqNum) + "_" + eventString)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0700)
	if err != nil {
		fmt.Printf("storeEvent OpenFile err = %v\n", err)
	}
	defer f.Close()
	serializedBytes, _ := json.Marshal(todoListCreated)
	Info.Println(fmt.Printf("storeEvent serializedBytes = %s\n", serializedBytes))
	_, err = f.Write(serializedBytes)
	if err != nil {
		fmt.Printf("storeEvent Write err = %v\n", err)
	}
}

func git(command string, args ...string) {
	exec.Command("git", append([]string{"-C", "storage", command}, args...)...).Run() }

func makeStorage() {
	_, err := os.Stat("storage")
	if err == nil { return }
	exec.Command("git", "init", "storage").Run()
	makeTheDirStruct()
	commit()
}
func commit() {
	git("add", ".")
	git("commit", "-m", "Initial commit")
}
func makeTheDirStruct() {
	os.MkdirAll("storage/event-stream", 0700)
	os.MkdirAll("storage/projections", 0700)
	os.OpenFile("storage/projections/.gitignore", os.O_CREATE, 0700)
	os.OpenFile("storage/event-stream/.gitignore", os.O_CREATE, 0700)
}
