package server

import (
	"database/sql"
	"log"
	"net/http"
)

func StartHTTPServer(db *sql.DB) {

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
