package server

import (
	"database/sql"
	"fmt"
	"forum/login"
	"forum/models"
	"forum/session"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

const (
	PATH            = "./templates/pages/"
	INDEX_PATH      = PATH + "index.html"
	LOGIN_PATH      = PATH + "login.html"
	REGISTER_PATH   = PATH + "register.html"
	PROFILE_PATH    = PATH + "profile.html"
	CREATEPOST_PATH = PATH + "create-post.html"
	POSTS_PATH      = PATH + "posts.html"
	COOKIES_PATH    = PATH + "cookies.html"
)

func Index(db *sql.DB, m *session.Manager) http.HandlerFunc {
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

		w, err = executeBuffer(t, w, nil)
		if err != nil {
			renderError(w, 500)
			return
		}
	}
}

// check that function it it work correctly ?
func Visitor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func LoginHandler(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			t, err := parseTemplates(LOGIN_PATH)
			if err != nil {
				renderError(w, 500)
				return
			}

			w, err = executeBuffer(t, w, nil)
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
				//http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
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

func RegisterHandler(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			t, err := parseTemplates(REGISTER_PATH)
			if err != nil {
				renderError(w, 500)
				return
			}

			w, err = executeBuffer(t, w, nil)
			if err != nil {
				renderError(w, 500)
				return
			}

		case http.MethodPost:

			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")
			passwordConfirm := r.FormValue("password_confirm")

			if username == "" || email == "" || password == "" {
				http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
				return
			}

			if password != passwordConfirm {
				renderError(w, 400)
				//http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
				return
			}

			err := login.CreateAccount(db, username, email, password)
			if err != nil {
				renderError(w, 400)
				//is this the right http error code ?
				//http.Error(w, "Erreur création du compte", http.StatusBadRequest)
				return
			}

			// 	userID, err := login.Login(db, email, password)
			// if err != nil {
			// 	http.Error(w, "Erreur récupération du compte", http.StatusInternalServerError)
			// 	return
			// }

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
			// http.SetCookie(w, &http.Cookie{
			// 	Name:     "pending_2fa",
			// 	Value:    "1",
			// 	HttpOnly: true,
			// 	Secure:   false,
			// 	SameSite: http.SameSiteStrictMode,
			// 	Path:     "/",
			// })
			// http.Redirect(w, r, "/2fa", http.StatusSeeOther)
			http.Redirect(w, r, "/cookies", http.StatusSeeOther)

		default:
			renderError(w, 405)
			return
		}
	}
}

func Profile(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, err := login.GetSessionCookie(r)
		if err != nil {
			renderError(w, 401)
			// http.Error(w, "Vous n'êtes pas connecté", http.StatusUnauthorized)
			return
		}

		t, err := parseTemplates(PROFILE_PATH)
		if err != nil {
			renderError(w, 500)
			return
		}

		//need to be optimised
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

		userSQL := db.QueryRow(`
	SELECT
		users.id,
		users.username,
		users.email,
		users.created_at,
		roles.name
	FROM users
	JOIN roles ON roles.id = users.roles_id
`)
		if err != nil {
			log.Println("Erreur récupération des infos de l'utilisateur")
			http.Error(w, "Erreur récupération des infos de l'utilisateur", http.StatusInternalServerError)
			return
		}

		var user models.User

		err = userSQL.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.Role,
		)
		log.Println(user)

		w, err = executeBuffer(t, w, user)
		if err != nil {
			log.Println("problème dans le rendu de la page profil")
			renderError(w, 500)
			return
		}
	}
}

func AcceptCookies(db *sql.DB, m *session.Manager) http.HandlerFunc {
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
			//http.Error(w, "Erreur création session", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
	INSERT INTO sessions (id, user_id, created_at, expires_at)
	VALUES (?, ?, datetime('now'), datetime('now', '+1 day'))
`, sessionID, userID)

		if err != nil {
			//http.Error(w, "Erreur session", http.StatusInternalServerError)
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
	if r.Method == http.MethodGet {
		t, err := parseTemplates(COOKIES_PATH)
		if err != nil {
			renderError(w, 500)
			return
		}

		w, err = executeBuffer(t, w, nil)
		if err != nil {
			renderError(w, 500)
			return
		}

	} else {
		renderError(w, 405)
		return
	}
}

func Logout(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			renderError(w, 405)
			return
		}
		sessionID, err := login.GetSessionCookie(r)

		if err == nil {
			db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		}

		login.DeleteSessionCookie(w)

		fmt.Fprintln(w, "Déconnexion réussie")
	}
}

// CreatePost crée un post pour l'utilisateur connecté.
func CreatePost(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			t, err := parseTemplates(CREATEPOST_PATH)
			if err != nil {
				renderError(w, 500)
				return
			}

			w, err = executeBuffer(t, w, nil)
			if err != nil {
				renderError(w, 500)
				return
			}

		case http.MethodPost:

			sessionID, err := login.GetSessionCookie(r)
			if err != nil {
				http.Error(w, "Vous devez être connecté", http.StatusUnauthorized)
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

			title := strings.TrimSpace(r.FormValue("title"))
			content := strings.TrimSpace(r.FormValue("content"))
			categories := r.Form["categories"]

			if title == "" {
				http.Error(w, "Le titre est obligatoire", http.StatusBadRequest)
				return
			}

			if content == "" {
				http.Error(w, "Le contenu est obligatoire", http.StatusBadRequest)
				return
			}

			categoryText := strings.Join(categories, ",")

			postID, err := uuid.NewV4()
			if err != nil {
				http.Error(w, "Erreur création du post", http.StatusInternalServerError)
				return
			}

			_, err = db.Exec(`
			INSERT INTO posts (id, user_id, title, content, categories, created_at)
			VALUES (?, ?, ?, ?, ?, datetime('now'))
		`, postID.String(), userID, title, content, categoryText)

			if err != nil {
				log.Println("ERREUR SQL READ POSTS :", err)
				http.Error(w, "Erreur récupération des posts", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/posts", http.StatusSeeOther)

		default:
			renderError(w, 405)
			return
		}
	}
}

// ReadPosts affiche les publications du forum.
func ReadPosts(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(`
	SELECT
		posts.id,
		posts.user_id,
		users.username,
		posts.title,
		posts.content,
		posts.categories,
		posts.created_at
	FROM posts
	JOIN users ON users.id = posts.user_id
	ORDER BY posts.created_at DESC
`)
		if err != nil {
			http.Error(w, "Erreur récupération des posts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)

		for rows.Next() {

			var post models.Post
			var categoryText string

			err := rows.Scan(
				&post.ID,
				&post.UserID,
				&post.Username,
				&post.Title,
				&post.Content,
				&categoryText,
				&post.CreatedAt,
			)

			if err != nil {
				http.Error(w, "Erreur lecture des posts", http.StatusInternalServerError)
				return
			}

			if categoryText != "" {
				post.Categories = strings.Split(categoryText, ",")
			}

			posts = append(posts, post)
		}

		if err := rows.Err(); err != nil {

			http.Error(w, "Erreur lecture des posts", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		t, err := parseTemplates(POSTS_PATH)
		if err != nil {
			renderError(w, 500)
			return
		}

		w, err = executeBuffer(t, w, posts)
		if err != nil {
			log.Println("ReadPosts : erreur de rendu dans l'execution du buffer")
			renderError(w, 500)
			return
		}
	}
}
