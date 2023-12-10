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

type Post struct {
	ID        int    `db:"id"`
	Title     string `db:"title"`
	Body      string `db:"body"`
	Author    string `db:"author"`
	CreatedAt string `db:"created_at"`
}


func main() {
	dbInit()
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

func dbConnect() *sqlx.DB {
	// SQLite3のデータベースに接続
	db, err := sqlx.Open("sqlite3", dbfileName)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func dbInit() {
	db := dbConnect()
	defer db.Close()

	// ブログポストテーブルを作成
	db.MustExec(createPostTable)
}

// ブログポストを作成
func dbInsert(title string, body string, author string, createdAt string) {
	db := dbConnect()
	defer db.Close()

	// ブログポストテーブルにデータを挿入
	db.MustExec(insertPostTable, title, body, author, createdAt)
}

// ブログポストを全件取得
func dbGetAll() []Post {
	db := dbConnect()
	defer db.Close()

	// ブログポストテーブルからデータを取得
	var posts []Post
	db.Select(&posts, selectPostTable)
	return posts
}

// ブログポストを1件取得
func dbGetOne(id int) Post {
	db := dbConnect()
	defer db.Close()

	// ブログポストテーブルからデータを取得
	var post Post
	db.Get(&post, selectPostTable+" WHERE id = ?", id)
	return post
}

// ブログポストを更新
func dbUpdate(id int, title string, body string, author string, createdAt string) {
	db := dbConnect()
	defer db.Close()

	// ブログポストテーブルのデータを更新
	db.MustExec(updatePostTable, title, body, author, createdAt, id)
}

// ブログポストを削除
func dbDelete(id int) {
	db := dbConnect()
	defer db.Close()

	// ブログポストテーブルのデータを削除
	db.MustExec(deletePostTable, id)
}

// ブログポストを全削除
func dbDeleteAll() {
	db := dbConnect()
	defer db.Close()

	// ブログポストテーブルのデータを全削除
	db.MustExec(deleteAllPostTable)
}

