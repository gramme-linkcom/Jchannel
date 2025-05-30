package main

import (
	"encoding/json" // JSONエンコードのために追加
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

// Post は掲示板の投稿を表す構造体です。
type Post struct {
	ID      int    `json:"id"`      // JSONタグを追加
	Content string `json:"content"` // JSONタグを追加
	Author  string `json:"author"`  // JSONタグを追加
	Date    string `json:"date"`    // JSONタグを追加
}

// Global variable to store posts in memory (for simplicity).
// 投稿をメモリに保存するためのグローバル変数（シンプルさのため）。
var (
	posts    []Post
	nextID   int
	postsMux sync.Mutex // 投稿へのアクセスを保護するためのミューテックス
)

// init はプログラム起動時に一度だけ実行されます。
func init() {
	posts = []Post{}
	nextID = 1
	// テスト用の初期投稿を追加
	posts = append(posts, Post{ID: nextID, Content: "こんにちは！Jちゃんねるへようこそ！", Author: "Jちゃんねる運営", Date: time.Now().Format("2006-01-02 15:04:05")})
	nextID++
	posts = append(posts, Post{ID: nextID, Content: "初めての投稿です！", Author: "テストユーザー", Date: time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05")})
	nextID++
}

// indexHandler は掲示板のトップページを表示します。
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// HTMLテンプレートをパースします。
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, "テンプレートの読み込みエラー: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 投稿データをテンプレートに渡してレンダリングします。
	// 初回ロード時はGo側でレンダリングしますが、その後の更新はJavaScriptで行います。
	postsMux.Lock()
	defer postsMux.Unlock()
	err = tmpl.Execute(w, posts) // 初期データをHTMLに埋め込む
	if err != nil {
		http.Error(w, "テンプレートのレンダリングエラー: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// createPostHandler は新しい投稿を処理します。
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTメソッドのみ許可されています。", http.StatusMethodNotAllowed)
		return
	}

	// フォームから投稿内容と著者を取得します。
	content := r.FormValue("content")
	author := r.FormValue("author")

	if content == "" || author == "" {
		http.Error(w, "投稿内容と著者は必須です。", http.StatusBadRequest)
		return
	}

	postsMux.Lock()
	defer postsMux.Unlock()

	// 新しい投稿を作成します。
	newPost := Post{
		ID:      nextID,
		Content: content,
		Author:  author,
		Date:    time.Now().Format("2006-01-02 15:04:05"),
	}
	posts = append(posts, newPost)
	nextID++

	// トップページにリダイレクトします。
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// getPostsAPIHandler は現在の投稿リストをJSON形式で返します。
func getPostsAPIHandler(w http.ResponseWriter, r *http.Request) {
	postsMux.Lock()
	defer postsMux.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, "JSONエンコードエラー: "+err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// ルーティングを設定します。
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/post", createPostHandler)
	http.HandleFunc("/api/posts", getPostsAPIHandler) // 新しいAPIエンドポイント

	log.Println("サーバーを起動しました: http://localhost:8080")
	// サーバーをポート8080で起動します。
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("サーバー起動エラー: ", err)
	}
}

//ファイルのネイティブアップロード
