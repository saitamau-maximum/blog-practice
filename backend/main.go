package main

import (
	"net/http"
	"fmt"
	"log"
	"text/template"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const Template = "../frontend"

const (
	dbfileName = "../db/db.sqlite3"
	// ブログポストテーブルを作成するSQL文
	createPostTable = `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		body TEXT,
		author TEXT,
		created_at TEXT
	)`

	// ブログポストテーブルにデータを挿入するSQL文
	insertPostTable = `INSERT INTO posts (title, body, author, created_at) VALUES (?, ?, ?, ?)`

	// ブログポストテーブルからデータを取得するSQL文
	selectPostTable = `SELECT * FROM posts`

	// ブログポストテーブルのデータを更新するSQL文
	updatePostTable = `UPDATE posts SET title = ?, body = ?, author = ?, created_at = ? WHERE id = ?`

	// ブログポストテーブルのデータを削除するSQL文
	deletePostTable = `DELETE FROM posts WHERE id = ?`

	// ブログポストテーブルのデータを全削除するSQL文
	deleteAllPostTable = `DELETE FROM posts`

)


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
