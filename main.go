package main

import (
	"net/http"
	"os"
	"os/exec"
	"html/template"
	"io/ioutil"
	"encoding/json"
)

type Order struct {
	Id int
	Date string
}
func listOrdersPageHandler(w http.ResponseWriter, r *http.Request) {
	orders := []Order{}
	file, err := ioutil.ReadFile("storage/projections/orders")
	if err != nil { json.Unmarshal(file, &orders)}
	PageTemplates.ExecuteTemplate(w, "orders.html", orders)
}

var PageTemplates = template.Must(template.ParseGlob("templates/*.html"))
func main() {
	makeStorage()
	http.HandleFunc("/listOrders", listOrdersPageHandler)
	http.ListenAndServe(":8080", nil)
}

func git(command string, args ...string) { exec.Command("git", append([]string{"-C", "storage", command}, args...)...) }

func makeStorage() {
	_, err := os.Stat("storage")
	if err == nil { return }
	exec.Command("git", "init", "storage")
	makeTheDirStruct()
	git("add", ".")
	git("commit", "-m='Initial commit'")
}
func makeTheDirStruct() {
	os.MkdirAll("storage/event-stream", 0700)
	os.MkdirAll("storage/projections", 0700)
	os.OpenFile("storage/projections/.gitignore", os.O_CREATE, 0700)
	os.OpenFile("storage/event-stream/.gitignore", os.O_CREATE, 0700)
}

