package server

import (
	"database/sql"
	"fmt"
	"forum/login"
	"html/template"
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

func parseTemplates(filenames string) (*template.Template, error) {
	path := "./templates/layouts/"
	t, err := template.ParseFiles(
		filenames,
		path+"head.html",
		path+"header.html",
		path+"footer.html")

	return t, err
}

func Index(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Need to add template parsing and execution
		// var buffer bytes.Buffer
		t, err := parseTemplates(INDEX_PATH)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// w.Header().Set("Content-Type:", "text/html; charset=utf-8")
		// w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
		// w.Write(buffer.Bytes())

		err = t.Execute(w, nil) //add the right variable as data to be rendered in the website
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			t, err := parseTemplates(LOGIN_PATH)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = t.Execute(w, nil) //add the right variable as data to be rendered in the website
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

		case http.MethodPost:

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

		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = t.Execute(w, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case http.MethodPost:

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

		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
	}
}

func Profile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := login.GetSessionCookie(r)
		if err != nil {
			http.Error(w, "Vous n'êtes pas connecté", http.StatusUnauthorized)
			return
		}

		t, err := parseTemplates(PROFILE_PATH)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		err = t.Execute(w, nil) //add the right variable as data to be rendered in the website
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
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
