package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})

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

		http.SetCookie(w, &http.Cookie{
			Name:     "pending_user_id",
			Value:    userID,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		http.Redirect(w, r, "/cookies", http.StatusSeeOther)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		err := CreateAccount(db, name, email, password)

		if err != nil {
			http.Error(w, "Erreur création du compte", http.StatusBadRequest)
			return
		}

		var userID string

		err = db.QueryRow(
			"SELECT id FROM users WHERE email = ?",
			email,
		).Scan(&userID)

		if err != nil {
			http.Error(w, "Erreur récupération utilisateur", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "pending_user_id",
			Value:    userID,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		http.Redirect(w, r, "/cookies", http.StatusSeeOther)
	})

	http.HandleFunc("/cookies", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		http.ServeFile(w, r, "templates/cookies.html")
	})

	http.HandleFunc("/accept-cookies", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		cookie, err := r.Cookie("pending_user_id")

		if err != nil {
			http.Error(w, "Aucune connexion en attente", http.StatusUnauthorized)
			return
		}

		userID := cookie.Value

		sessionID, err := CreateSessionCookie(w)

		if err != nil {
			http.Error(w, "Erreur création session", http.StatusInternalServerError)
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

		http.SetCookie(w, &http.Cookie{
			Name:     "pending_user_id",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		http.Redirect(w, r, "/profil", http.StatusSeeOther)
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
