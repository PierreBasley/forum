package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

func StartHTTPServer(db *sql.DB) {
	// SQL scripts should be read directly from a file and executed, not
	// hardcoded
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		expires_at DATETIME NOT NULL
	)
`)

	if err != nil {
		fmt.Println("Erreur création sessions :", err)
		return
	}

	http.HandleFunc("/", Index(db))
	http.HandleFunc("GET /login", LoginPage)
	http.HandleFunc("GET /visitor", Visitor)
	http.HandleFunc("GET /2fa", TwoFactorPage)
	http.HandleFunc("POST /login", LoginHandler(db))
	http.HandleFunc("POST /2fa", TwoFactorHandler(db))
	http.HandleFunc("POST /register", Register(db))
	http.HandleFunc("GET /logout", Logout(db))

	http.HandleFunc("GET /cookies", Cookies)
	http.HandleFunc("POST /accept-cookies", AcceptCookies(db))

	http.HandleFunc("GET /profile", Profile(db))
	http.HandleFunc("POST /posts", CreatePost(db))
	http.HandleFunc("GET /posts", ReadPosts(db))

	http.HandleFunc("GET /create-post", CreatePostPage(db))

	greet()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func greet() {
	log.Print("Serveur HTTP lancé sur http://localhost:8080")
}
