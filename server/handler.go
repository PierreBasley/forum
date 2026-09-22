package server

import (
	"database/sql"
	"fmt"
	"forum/login"
	"net/http"
	"time"
)

func Index(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, "templates/index.html")
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		userID, err := login.Login(db, email, password)

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
	}
}

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		passwordConfirm := r.FormValue("password_confirm")

		if password != passwordConfirm {
			http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
			return
		}

		err := login.CreateAccount(db, name, email, password)

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
	}
}

func Profile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := login.GetSessionCookie(r)

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
	}
}

func AcceptCookies(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		sessionID, err := login.CreateSessionCookie(w)

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

		http.Redirect(w, r, "/create-post", http.StatusSeeOther)
	}
}

func Cookies(w http.ResponseWriter, r *http.Request) {
	// Again, should use templates
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "templates/cookies.html")
}

func Logout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := login.GetSessionCookie(r)

		if err == nil {
			db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		}

		login.DeleteSessionCookie(w)

		fmt.Fprintln(w, "Déconnexion réussie")
	}
}

func CreatePostPage(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "templates/create-post.html")
}
