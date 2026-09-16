package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

func startHTTPServer(db *sql.DB) {

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

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		userID, err := Login(db, email, password)

		if err != nil {
			http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
			return
		}

		sessionID, err := CreateSessionCookie(w)

		if err != nil {
			http.Error(w, "Erreur session", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
			INSERT INTO sessions (id, user_id, expires_at)
			VALUES (?, ?, datetime('now', '+1 day'))
		`, sessionID, userID)

		if err != nil {
			http.Error(w, "Erreur session", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Connexion réussie")
	})

	http.HandleFunc("/profil", func(w http.ResponseWriter, r *http.Request) {

		sessionID, err := GetSessionCookie(r)

		if err != nil {
			http.Error(w, "Vous n'êtes pas connecté", http.StatusUnauthorized)
			return
		}

		var userID string

		err = db.QueryRow(`
			SELECT user_id
			FROM sessions
			WHERE id = ?
			AND expires_at > datetime('now')
		`, sessionID).Scan(&userID)

		if err != nil {
			http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
			return
		}

		fmt.Fprintf(w, "Vous êtes connecté. User ID : %s\n", userID)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {

		sessionID, err := GetSessionCookie(r)

		if err == nil {
			db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		}

		DeleteSessionCookie(w)

		fmt.Fprintln(w, "Déconnexion réussie")
	})

	fmt.Println("Serveur HTTP lancé sur http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}
