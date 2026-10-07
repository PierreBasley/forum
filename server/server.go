package server

import (
	"database/sql"
	"log"
	"net/http"
)

func StartHTTPServer(db *sql.DB) {

	http.HandleFunc("/", Index(db))
	http.HandleFunc("/login", LoginHandler(db))
	//http.HandleFunc("GET /login", LoginPage)
	http.HandleFunc("GET /visitor", Visitor)
	//http.HandleFunc("GET /2fa", TwoFactorPage)
	//http.HandleFunc("POST /login", LoginHandler(db))
	http.HandleFunc("GET /forgot-password", ForgotPasswordPage)
	http.HandleFunc("POST /forgot-password", ForgotPasswordHandler(db))
	http.HandleFunc("GET /reset-password", ResetPasswordPage)
	http.HandleFunc("POST /reset-password", ResetPasswordHandler(db))
	//http.HandleFunc("POST /2fa", TwoFactorHandler(db))
	//http.HandleFunc("POST /2fa/resend", ResendTwoFactorHandler(db))
	http.HandleFunc("/register", RegisterHandler(db))
	http.HandleFunc("GET /profile", Profile(db))
	http.HandleFunc("GET /logout", Logout(db))
	http.HandleFunc("/cookies", Cookies)
	http.HandleFunc("/accept-cookies", AcceptCookies(db))

	http.HandleFunc("/create-post", CreatePost(db))
	http.HandleFunc("/posts", ReadPosts(db))

	greet()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func greet() {
	log.Print("Serveur HTTP lancé sur http://localhost:8080")
}
