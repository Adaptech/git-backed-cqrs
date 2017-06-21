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
	http.ListenAndServe(":8080", nil)
}
func createTodoList(writer http.ResponseWriter, request *http.Request) {
	createTodoCommand := &TodoListCreate{
		Message {
			pseudo_uuid(),
		},
		"test",
	}
	handleCreateTodoList(createTodoCommand)
}
func handleCreateTodoList(todoListCreate *TodoListCreate) {
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
	dirs, _ := ioutil.ReadDir(streamDir)
	fileName := string(len(dirs) + 1) + "_" + reflect.TypeOf(todoListCreated).Name()
	serializedBytes, _ := json.Marshal(todoListCreated)
	ioutil.WriteFile(path.Join(streamDir, fileName), serializedBytes, 600 )
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
