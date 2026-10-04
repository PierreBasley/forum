package server

import (
	"database/sql"
	"fmt"
	"forum/login"
	"net/http"
	"time"
)

const (
	PATH          = "./templates/pages/"
	INDEX_PATH    = PATH + "index.html"
	LOGIN_PATH    = PATH + "login.html"
	REGISTER_PATH = PATH + "register.html"
	PROFILE_PATH  = PATH + "profile.html"
)

func Index(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// YF: Need to add template parsing and execution.
		// PB : Done + integration of the buffer

		if r.URL.Path != "/" {
			renderError(w, 404)
			return
		}

		t, err := parseTemplates(INDEX_PATH)
		if err != nil {
			renderError(w, 500)
			return
		}

		w, err = executeBuffer(t, w)
		if err != nil {
			renderError(w, 500)
			return
		}
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			t, err := parseTemplates(LOGIN_PATH)
			if err != nil {
				renderError(w, 500)
				return
			}

			w, err = executeBuffer(t, w)
			if err != nil {
				renderError(w, 500)
				return
			}

		case http.MethodPost:

			email := r.FormValue("email")
			password := r.FormValue("password")

			userID, err := login.Login(db, email, password)
			if err != nil {
				renderError(w, 401)
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

		default:
			renderError(w, 405)
			return
		}
	}
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			t, err := parseTemplates(REGISTER_PATH)
			if err != nil {
				renderError(w, 500)
				return
			}

			w, err = executeBuffer(t, w)
			if err != nil {
				renderError(w, 500)
				return
			}

		case http.MethodPost:

			name := r.FormValue("name")
			email := r.FormValue("email")
			password := r.FormValue("password")
			passwordConfirm := r.FormValue("password_confirm")

			if password != passwordConfirm {
				renderError(w, 400)
				//is this the right http error code ?
				//http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
				return
			}

			err := login.CreateAccount(db, name, email, password)
			if err != nil {
				renderError(w, 400)
				//is this the right http error code ?
				//http.Error(w, "Erreur création du compte", http.StatusBadRequest)
				return
			}

			var userID string

			err = db.QueryRow(
				"SELECT id FROM users WHERE email = ?",
				email,
			).Scan(&userID)

			if err != nil {
				renderError(w, 500)
				//http.Error(w, "Erreur récupération utilisateur", http.StatusInternalServerError)
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

		default:
			renderError(w, 405)
			return
		}
	}
}

func Profile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := login.GetSessionCookie(r)
		if err != nil {
			renderError(w, 401)
			return
		}

		t, err := parseTemplates(REGISTER_PATH)
		if err != nil {
			renderError(w, 500)
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

		w, err = executeBuffer(t, w)
		if err != nil {
			renderError(w, 500)
			return
		}
	}
}

func AcceptCookies(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			renderError(w, 405)
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
			renderError(w, 500)
			return
		}

		_, err = db.Exec(`
	INSERT INTO sessions (id, user_id, expires_at)
	VALUES (?, ?, datetime('now', '+1 day'))
`, sessionID, userID)

		if err != nil {
			renderError(w, 500)
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

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}

func Cookies(w http.ResponseWriter, r *http.Request) {
	// Again, should use templates
	if r.Method != http.MethodGet {
		renderError(w, 405)
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
