package server

import (
	"database/sql"
	"fmt"
	"forum/session"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	ENV_PORT_NUMBER = "PORT_CONTAINER"
	DEFAULT_PORT = 8080
)

func StartHTTPServer(db *sql.DB, m *session.Manager) {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", Index(db, m))
	http.HandleFunc("/login", LoginHandler(db, m))
	http.HandleFunc("/forgot-password", ForgotPasswordHandler(db, m))
	http.HandleFunc("/reset-password", ResetPasswordHandler(db, m))
	http.HandleFunc("/register", RegisterHandler(db, m))
	http.HandleFunc("/profile", Profile(db, m))
	http.HandleFunc("/logout", Logout(db, m))
	http.HandleFunc("/cookies", Cookies)
	http.HandleFunc("/create-post", CreatePost(db, m))
	http.HandleFunc("/posts", ReadPosts(db, m))
	http.HandleFunc("/posts/{post_id}", ReadPost(db, m))
	http.HandleFunc("/comment", Comment(db, m))
	http.HandleFunc("/like", LikeHandler(db, m))

	greet()

	portNumber := DEFAULT_PORT
	if os.Getenv(ENV_PORT_NUMBER) != "" {
		envPort, err := strconv.Atoi(os.Getenv(ENV_PORT_NUMBER))
		if err == nil {
			portNumber = envPort
		}
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNumber), nil))
}

func greet() {
	portNumber := DEFAULT_PORT
	if os.Getenv(ENV_PORT_NUMBER) != "" {
		envPort, err := strconv.Atoi(os.Getenv(ENV_PORT_NUMBER))
		if err == nil {
			portNumber = envPort
		}
	}
	log.Printf("Serveur HTTP lancé sur http://localhost:%d", portNumber)
}
