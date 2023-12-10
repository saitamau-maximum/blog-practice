package main

import (
	"net/http"
	"fmt"
	"log"
	"text/template"
)

const Template = "../frontend"



func main() {
	http.HandleFunc("/", IndexHandler)
	fmt.Println("http://localhost:8080 で起動しています...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(Template + "/index.html")
	// {{.Title}}と{{.Body}}に対して、それぞれ"Hello, World"と"こんにちは、世界"を埋め込む
	t.Execute(w, struct {
		Title string
		Body  string
	}{
		Title: "Hello, World",
		Body:  "こんにちは、世界",
	})
	if err != nil {
		log.Fatal(err)
	}
	
}
