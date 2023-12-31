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

const templatePath = "./templates"
const layoutPath = templatePath + "/layout.html"
const publicPath = "./public"

const (
	dbPath = "./db.sqlite3"
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
	CreatedAt int64  `db:"created_at"`
}

var (
	db *sqlx.DB

	funcDate = template.FuncMap{
		"date": func(t int64) string {
			return time.Unix(t, 0).Format("2006-01-02 15:04:05")
		},
	}

	indexTemplate = template.Must(template.New("layout.html").Funcs(funcDate).ParseFiles(layoutPath, templatePath+"/index.html"))

	postTemplate = template.Must(template.New("layout.html").Funcs(funcDate).ParseFiles(layoutPath, templatePath+"/post.html"))

	createTemplate = template.Must(template.New("layout.html").ParseFiles(layoutPath, templatePath+"/create.html"))
)

func main() {
	// データベースに接続
	db = dbConnect()
	defer db.Close()
	// データベースの初期化
	err := initDB()
	if err != nil {
		// データベースの初期化に失敗した場合は終了
		log.Fatal(err)
	}
	http.HandleFunc("/", IndexHandler)
	// blog表示用のハンドラーを追加　/blog/idの形式でアクセスされた場合にblogHandlerが呼ばれる /blog/newの形式でアクセスされた場合にcreatePostHandlerが呼ばれる
	http.HandleFunc("/post/", BlogHandler)
	http.HandleFunc("/post/new", CreatePostHandler)

	// cssファイルを配信
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(publicPath+"/css"))))

	fmt.Println("http://localhost:8080 で起動しています...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	//ブログポストを全件取得
	posts := getAllPosts()
	//ブログポストをテンプレートに渡す
	indexTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"Posts": posts,
		"PageTitle": "ブログポスト一覧",
	})

}

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	// /blog/idの形式でアクセスされた場合にidを取得
	id := r.URL.Path[len("/post/"):]
	// idをint型に変換
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Print(err)
		return
	}
	// ブログポストを1件取得
	post, err := getPostById(idInt)
	if err != nil {
		log.Print(err)
		return
	}
	// ブログポストをテンプレートに渡す
	postTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"Title":     post.Title,
		"PageTitle": post.Title,
		"Body":      post.Body,
		"Author":    post.Author,
		"CreatedAt": post.CreatedAt,
	})
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// GETリクエストの場合はテンプレートを表示
		createTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
			"PageTitle": "ブログポスト作成",
		})
	} else if r.Method == "POST" {
		// POSTリクエストの場合はブログポストを作成
		title := r.FormValue("title")
		body := r.FormValue("body")
		author := r.FormValue("author")
		createdAt := time.Now().Unix()
		// フォームに空の項目がある場合はエラーを返す
		if title == "" || body == "" || author == "" {
			log.Print("フォームに空の項目があります")
			createTemplate.ExecuteTemplate(w, "layout.html", map[string]interface{}{
				"Message": "フォームに空の項目があります",
			})
			return
		}
		id, err := insertPost(title, body, author, createdAt)
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

func initDB() error {
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
func insertPost(title string, body string, author string, createdAt int64) (int64, error) {
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
func getAllPosts() []Post {
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
func getPostById(id int) (Post, error) {
	// ブログポストテーブルからデータを取得
	var post Post
	err := db.Get(&post, selectPostByIdQuery, id)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return Post{}, err
	}
	return post, nil
}

// ブログポストを更新
func updatePost(id int, title string, body string, author string, createdAt int64) error {
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
func deletePostById(id int) error {
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
func deleteAllPosts() error {
	// ブログポストテーブルのデータを全削除
	_, err := db.Exec(deleteAllPostsQuery)
	if err != nil {
		log.Print(err)
		// InternalServerErrorを返す
		return err
	}
	return nil
}
