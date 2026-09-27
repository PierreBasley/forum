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
	http.HandleFunc("/login", LoginHandler(db))
	http.HandleFunc("/register", RegisterHandler(db))
	http.HandleFunc("/profile", Profile(db))
	http.HandleFunc("GET /logout", Logout(db))

	http.HandleFunc("GET /cookies", Cookies)
	http.HandleFunc("GET /accept-cookies", AcceptCookies(db))

	greet()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func greet() {
	log.Print("Serveur HTTP lancé sur http://localhost:8080")
}
