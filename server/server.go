package server

import (
	"database/sql"
	"forum/session"
	"log"
	"net/http"
)

func StartHTTPServer(db *sql.DB, m *session.Manager) {

	http.HandleFunc("/", Index(db, m))
	http.HandleFunc("/login", LoginHandler(db, m))
	//http.HandleFunc("GET /login", LoginPage)
	http.HandleFunc("GET /visitor", Visitor)
	//http.HandleFunc("GET /2fa", TwoFactorPage)
	//http.HandleFunc("POST /login", LoginHandler(db))
	//http.HandleFunc("GET /forgot-password", ForgotPasswordPage)
	//http.HandleFunc("POST /forgot-password", ForgotPasswordHandler(db))
	//http.HandleFunc("GET /reset-password", ResetPasswordPage)
	//http.HandleFunc("POST /reset-password", ResetPasswordHandler(db))
	//http.HandleFunc("POST /2fa", TwoFactorHandler(db))
	//http.HandleFunc("POST /2fa/resend", ResendTwoFactorHandler(db))
	http.HandleFunc("/register", RegisterHandler(db, m))
	http.HandleFunc("GET /profile", Profile(db, m))
	http.HandleFunc("GET /logout", Logout(db, m))
	http.HandleFunc("/cookies", Cookies)
	http.HandleFunc("/accept-cookies", AcceptCookies(db, m))

	http.HandleFunc("/create-post", CreatePost(db, m))
	http.HandleFunc("/posts", ReadPosts(db, m))

	greet()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func greet() {
	log.Print("Serveur HTTP lancé sur http://localhost:8080")
}
