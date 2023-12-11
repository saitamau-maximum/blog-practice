package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const templatePath = "../frontend"

const (
	dbPath = "../db/db.sqlite3"
	//ブログポストテーブルを作成するSQL文
	createPostTableQuery = `CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		body TEXT,
		author TEXT,
		created_at INTEGER
	)`

	// ブログポストテーブルにデータを挿入するSQL文
	insertPostQuery = `INSERT INTO posts (title, body, author, created_at) VALUES (?, ?, ?, ?)`

	// ブログポストテーブルからデータを取得するSQL文
	selectAllPostsQuery = `SELECT * FROM posts`

	// ブログポストテーブルのデータを更新するSQL文
	updatePostQuery = `UPDATE posts SET title = ?, body = ?, author = ?, created_at = ? WHERE id = ?`

	// ブログポストテーブルのデータを削除するSQL文
	deletePostQuery = `DELETE FROM posts WHERE id = ?`

	// ブログポストテーブルのデータを全削除するSQL文
	deleteAllPostsQuery = `DELETE FROM posts`

	selectPostByIdQuery = `SELECT * FROM posts WHERE id = ?`
)

type Post struct {
	ID        int    `db:"id"`
	Title     string `db:"title"`
	Body      string `db:"body"`
	Author    string `db:"author"`
	CreatedAt int64 `db:"created_at"`
}

var (
	db *sqlx.DB

	funcDate = template.FuncMap{
		"date": func(t int64) string {
			return time.Unix(t, 0).Format("2006-01-02 15:04:05")
		},
	}

	indexTemplate = template.Must(template.New("index.html").Funcs(funcDate).ParseFiles(templatePath + "/index.html"))

	postTemplate = template.Must(template.New("post.html").Funcs(funcDate).ParseFiles(templatePath + "/post.html"))

	createTemplate = template.Must(template.New("create.html").ParseFiles(templatePath + "/create.html"))
)

func main() {
	// データベースに接続
	db = dbConnect()
	defer db.Close()
	// データベースの初期化
	dbInit()
	http.HandleFunc("/", IndexHandler)
	// // blog表示用のハンドラーを追加　/blog/idの形式でアクセスされた場合にblogHandlerが呼ばれる /blog/createの形式でアクセスされた場合にcreatePostHandlerが呼ばれる
	http.HandleFunc("/post/", BlogHandler)
	http.HandleFunc("/post/create", CreatePostHandler)

	fmt.Println("http://localhost:8080 で起動しています...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// ブログポストを全件取得
	posts := dbGetAll()
	// ブログポストをテンプレートに渡す
	indexTemplate.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Posts": posts,
	})

}

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	// /blog/idの形式でアクセスされた場合にidを取得
	id := r.URL.Path[len("/post/"):]
	// idをint型に変換
	idInt, err := strconv.Atoi(id)
	println(idInt)
	if err != nil {
		log.Fatal(err)
	}
	// ブログポストを1件取得
	post := dbGetOne(idInt)
	// ブログポストをテンプレートに渡す
	postTemplate.ExecuteTemplate(w, "post.html", map[string]interface{}{
		"Title":     post.Title,
		"Body":      post.Body,
		"Author":    post.Author,
		"CreatedAt": post.CreatedAt,
	})
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// GETリクエストの場合はテンプレートを表示
		createTemplate.Execute(w, nil)
	} else if r.Method == "POST" {
		// POSTリクエストの場合はブログポストを作成
		title := r.FormValue("title")
		body := r.FormValue("body")
		author := r.FormValue("author")
		createdAt := time.Now().Unix()
		id, err := dbInsert(title, body, author, createdAt)
		if err != nil {
			log.Print(err)
			return 
		}
		// 作成したブログポストを表示
		http.Redirect(w, r, "/post/"+strconv.FormatInt(id, 10), 301)
	}
}

func dbConnect() *sqlx.DB {
	// SQLite3のデータベースに接続
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func dbInit() error {
	// ブログポストテーブルを作成
	_, err := db.Exec(createPostTableQuery)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return err
	}
	return nil
}

// ブログポストを作成
func dbInsert(title string, body string, author string, createdAt int64) (int64, error) {
	// ブログポストテーブルにデータを挿入　last_insert_rowid()で最後に挿入したデータのIDを取得
	result, err := db.Exec(insertPostQuery, title, body, author, createdAt)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return 0, err
	}
	return id, nil
}

// ブログポストを全件取得
func dbGetAll() []Post {
	// ブログポストテーブルからデータを取得
	var posts []Post
	db.Select(&posts, selectAllPostsQuery)
	// 何も取得できなかった場合は空のスライスを返す
	if len(posts) == 0 {
		return []Post{}
	}
	return posts
}

// ブログポストを1件取得
func dbGetOne(id int) Post {
	// ブログポストテーブルからデータを取得
	var post Post
	db.Get(&post, selectPostByIdQuery, id)
	return post
}

// ブログポストを更新
func dbUpdate(id int, title string, body string, author string, createdAt string) error {
	// ブログポストテーブルのデータを更新
	_, err := db.Exec(updatePostQuery, title, body, author, createdAt, id)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return err
	}
	return nil
}

// ブログポストを削除
func dbDelete(id int) error {
	// ブログポストテーブルのデータを削除
	_, err := db.Exec(deletePostQuery, id)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return err
	}
	return nil
}

// ブログポストを全削除
func dbDeleteAll() error {
	// ブログポストテーブルのデータを全削除
	_, err := db.Exec(deleteAllPostsQuery)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return err
	}
	return nil
}
