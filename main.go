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
	"strconv"
)

type TodoList struct {
	Name string
}
type TodoListCreated struct {
	Identity
	Name string
}
func (event TodoListCreated) GetId() string { return event.Id }

type TodoListItemCreated struct {
	Identity
	Name string
}
type Identity struct {
	Id string
}
type TodoListCreate struct {
	Identity
	Name string
}
type Message interface {
	GetId() string
}
func listTodoListsPageHandler(w http.ResponseWriter, r *http.Request) {
//	todoLists := []TodoList{}
	var todoLists []TodoList
	file, err := ioutil.ReadFile("storage/projections/todoLists")
	if err != nil {
		fmt.Println("listTodoListsPageHandler error: ", err)
	} else {
		fmt.Printf("listTodoListsPageHandler file = %s\n", file)
		err = json.Unmarshal(file, &todoLists)
		if err != nil {
			fmt.Println("listTodoListsPageHandler Unmarshal error: ", err)
		} else {
			fmt.Printf("listTodoListsPageHandler unmarshalled = %#v\n", todoLists)
		}
	}
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

	fmt.Println("Now listening on http://localhost:8080 ...")

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
		Identity{
			pseudo_uuid(),
		},
		request.FormValue("name"),
	}
	handleCreateTodoList(createTodoCommand)
}
func handleCreateTodoList(todoListCreate *TodoListCreate) {
	fmt.Printf("handleCreateTodoList Id = %v Name = %v\n", todoListCreate.Id, todoListCreate.Name)
	todoListCreated := TodoListCreated{
		Identity{
			todoListCreate.Id,
		},
		todoListCreate.Name,
	}
	storeEvent(todoListCreated)
}
func storeEvent(event Message) {
	//TODO: wrap all of this so if anything fails we reset or stash the working directory
	name := UpdateStream(event)
	UpdateStreamIndex(name)
	UpdateReadModels(event)
	//TODO: then add and commit in git
	git("add", ".")
	git("commit", "-m",event.GetId())
}
func UpdateStreamIndex(eventName string) {
	f, err := os.OpenFile("storage/event-stream/index", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0700)
	if err != nil {
		fmt.Printf("UpdateStreamIndex OpenFile err = %v\n", err)
	}
	defer f.Close()
	_, err = f.WriteString(eventName+"\n")
	if err != nil {
		fmt.Printf("UpdateStreamIndex Write err = %v\n", err)
	}
}

// This is your pub/sub.. the kafka replacement
func UpdateReadModels(event interface{}) {
	UpdateListOfTodos(event)
	UpdateTodoDetails(event)
	UpdateCountOfTodoLists(event)
}

// An example of a read model
func UpdateCountOfTodoLists(event interface{}) {
	count := 0
	fileStoredCount, err := ioutil.ReadFile("storage/projections/TodoListsCount")
	if err != nil {
		ioutil.WriteFile("storage/projections/TodoListsCount", []byte("1"), 0700)
		return
	}
	count, _ = strconv.Atoi(string(fileStoredCount))
	ioutil.WriteFile("storage/projections/TodoListsCount", []byte(strconv.Itoa(count + 1)), 0700)
}
func UpdateTodoDetails(event interface{}) {

}
func UpdateListOfTodos(event interface{}) {
	fn := "storage/projections/todoLists"
	fi, err := os.Stat(fn)
	fmt.Println("UpdateListOfTodos Stat err = ", err)
	if err != nil {
		f, err := os.Create(fn)
		if err != nil {
			fmt.Println("UpdateListOfTodos Create err = ", err)
		}
		err = f.Chmod(0700)
		if err != nil {
			fmt.Println("UpdateListOfTodos Chmod err = ", err)
		}
		_, err = f.Write([]byte("["))
		if err != nil {
			fmt.Println("UpdateListOfTodos 1st Write err = ", err)
		}
		f.Close()
	}
	fi, err = os.Stat(fn)
	if fi.Size() > 1 {
		fmt.Println("UpdateListOfTodos file size = ", fi.Size())
		f, err := os.OpenFile("storage/projections/todoLists", os.O_RDWR|os.O_APPEND, 0700)
		if err != nil {
			fmt.Printf("UpdateListOfTodos 1st OpenFile err = %v\n", err)
		}
		err = f.Truncate(fi.Size()-1)
		if err != nil {
			fmt.Printf("UpdateListOfTodos Truncate err = %v\n", err)
		}
		f.Close()
	}
	f, err := os.OpenFile("storage/projections/todoLists", os.O_RDWR|os.O_APPEND, 0700)
	defer f.Close()
	if fi.Size() > 1 {
		_, err = f.Write([]byte(","))
		if err != nil {
			fmt.Println("UpdateListOfTodos 2nd Write err = ", err)
		}
	}
	serializedBytes, _ := json.Marshal(event)
	Info.Println(fmt.Printf("UpdateListOfTodos serializedBytes = %s\n", serializedBytes))
	_, err = f.Write(serializedBytes)
	if err != nil {
		fmt.Println("UpdateListOfTodos 3rd Write err = ", err)
	}
	_, err = f.Write([]byte("]"))
	if err != nil {
		fmt.Println("UpdateListOfTodos 4th Write err = ", err)
	}
}
// how event streams are stored
func UpdateStream(event Message) (string) {
	streamDir := path.Join("storage/event-stream", event.GetId())
	Info.Println(fmt.Printf("storeEvent streamDir = %v\n", streamDir))
	dirs, _ := ioutil.ReadDir(streamDir)
	if len(dirs) == 0 {
		os.Mkdir(streamDir, 0755)
	}
	Info.Println(fmt.Printf("storeEvent dirs = %v len = %v\n", dirs, len(dirs)))
	seqNum := len(dirs) + 1
	cmdStrs := strings.Split(fmt.Sprintf("%s", reflect.TypeOf(event)), ".")
	eventString := cmdStrs[len(cmdStrs)-1]
	fileName := path.Join(streamDir, fmt.Sprintf("%06d", seqNum)+"_"+eventString)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0700)
	if err != nil {
		fmt.Printf("storeEvent OpenFile err = %v\n", err)
	}
	defer f.Close()
	serializedBytes, _ := json.Marshal(event)
	Info.Println(fmt.Printf("storeEvent serializedBytes = %s\n", serializedBytes))
	_, err = f.Write(serializedBytes)
	if err != nil {
		fmt.Printf("storeEvent Write err = %v\n", err)
	}
	return fileName
}

func git(command string, args ...string) {
	exec.Command("git", append([]string{"-C", "storage", command}, args...)...).Run()
}

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
