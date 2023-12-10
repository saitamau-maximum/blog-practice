package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"strconv"
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

	// last_insert_rowid()を使って最後に挿入したデータのIDを取得するSQL文
	lastInsertID = `SELECT last_insert_rowid() AS id`

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
	// blog表示用のハンドラーを追加　/blog/idの形式でアクセスされた場合にblogHandlerが呼ばれる
	http.HandleFunc("/post/", BlogHandler)
	// create-post用のハンドラーを追加
	http.HandleFunc("/create-post", CreatePostHandler)
	fmt.Println("http://localhost:8080 で起動しています...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(Template + "/index.html")
	// ブログポストを全件取得
	posts := dbGetAll()
	if err != nil {
		log.Fatal(err)
	}
	// ブログポストをテンプレートに渡す
	t.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Posts": posts,
	})

}

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(Template + "/post.html")
	if err != nil {
		log.Fatal(err)
	}
	// /blog/idの形式でアクセスされた場合にidを取得
	id := r.URL.Path[len("/blog/"):]
	// idをint型に変換
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Fatal(err)
	}
	// ブログポストを1件取得
	post := dbGetOne(idInt)
	// ブログポストをテンプレートに渡す
	t.ExecuteTemplate(w, "post.html", map[string]interface{}{
		"Title":     post.Title,
		"Body":      post.Body,
		"Author":    post.Author,
		"CreatedAt": post.CreatedAt,
	})
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(Template + "/create-post.html")
	if err != nil {
		log.Fatal(err)
	}
	if r.Method == "GET" {
		// GETリクエストの場合はテンプレートを表示
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		// POSTリクエストの場合はブログポストを作成
		title := r.FormValue("title")
		body := r.FormValue("body")
		author := r.FormValue("author")
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		dbInsert(title, body, author, createdAt)
		// 最後に挿入したデータのIDを取得
		var id int
		dbGetLastInsertID := dbConnect()
		defer dbGetLastInsertID.Close()
		dbGetLastInsertID.Get(&id, lastInsertID)
		// 作成したブログポストを表示
		http.Redirect(w, r, "/blog/"+strconv.Itoa(id), http.StatusFound)
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
	// 何も取得できなかった場合は空のスライスを返す
	if len(posts) == 0 {
		return []Post{}
	}
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

