package server

import (
	"database/sql"
	"forum/session"
	"log"
	"net/http"
)

func StartHTTPServer(db *sql.DB, m *session.Manager) {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", Index(db, m))
	http.HandleFunc("/login", LoginHandler(db, m))
	http.HandleFunc("/forgot-password", ForgotPasswordHandler(db, m))
	http.HandleFunc("/reset-password", ResetPasswordHandler(db, m))
	//http.HandleFunc("POST /2fa", TwoFactorHandler(db))
	//http.HandleFunc("POST /2fa/resend", ResendTwoFactorHandler(db))
	http.HandleFunc("/register", RegisterHandler(db, m))
	http.HandleFunc("/profile", Profile(db, m))
	http.HandleFunc("/logout", Logout(db, m))
	http.HandleFunc("/cookies", Cookies)
	http.HandleFunc("/create-post", CreatePost(db, m))
	http.HandleFunc("/posts", ReadPosts(db, m))

	greet()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func greet() {
	log.Print("Serveur HTTP lancé sur http://localhost:8080")
}
