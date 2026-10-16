package server

import (
	"database/sql"
	"errors"
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
	FORGOT_PASSWORD = PATH + "forgot-password.html"
	RESET_PASSWORD  = PATH + "reset-password.html"
	PROFILE_PATH    = PATH + "profile.html"
	CREATEPOST_PATH = PATH + "create-post.html"
	POSTS_PATH      = PATH + "posts.html"
	COOKIES_PATH    = PATH + "cookies.html"
	LOGOUT_PATH     = PATH + "logout.html"

	COOKIE_NAME = "session_ID"
)

func Index(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" {
			renderError(w, 404, nil)
			return
		}

		var userData session.Authentified

		_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
		if err != nil {
			renderError(w, 500, nil)
			return
		}
		if ok == true {
			userData.IsActive = true
		}

		t, err := parseTemplates(INDEX_PATH)
		if err != nil {
			renderError(w, 500, nil)
			return
		}

		pageData := models.PageData{Auth: userData}
		w, err = executeBuffer(t, w, pageData)
		if err != nil {
			renderError(w, 500, nil)
			return
		}
	}
}

func LoginHandler(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:

			var userData session.Authentified

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == true {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Vous êtes déjà connecté"), AllowedMethods: nil})
				return
			}

			pageData := models.PageData{Auth: userData}

			t, err := parseTemplates(LOGIN_PATH)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, pageData)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

		case http.MethodPost:

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == true {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Vous êtes déjà connecté"), AllowedMethods: nil})
				return
			}

			email := r.FormValue("email")
			password := r.FormValue("password")

			userID, err := login.Login(db, email, password)
			if err != nil {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Email ou mot de passe incorrect"), AllowedMethods: nil})
				return
			}

			sessionID, expiresAt, err := m.CreationSession(userID)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			cookieUser := login.CreateCookie(sessionID, expiresAt, COOKIE_NAME)
			http.SetCookie(w, cookieUser)

			http.Redirect(w, r, "/cookies", http.StatusSeeOther)

		default:
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET", "POST"}})
			return
		}
	}
}

func RegisterHandler(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:

			var userData session.Authentified

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == true {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Vous êtes déjà connecté"), AllowedMethods: nil})
				return
			}

			pageData := models.PageData{Auth: userData}

			t, err := parseTemplates(REGISTER_PATH)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, pageData)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

		case http.MethodPost:

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == true {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Vous êtes déjà connecté"), AllowedMethods: nil})
				return
			}

			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")
			passwordConfirm := r.FormValue("password_confirm")

			if username == "" || email == "" || password == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Tous les champs sont obligatoires"), AllowedMethods: nil})
				return
			}

			if password != passwordConfirm {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Les mots de passe ne correspondent pas"), AllowedMethods: nil})
				return
			}

			if login.IsValidPassword(password) == false {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le mot de passe doit contenir au minimum 8 caractères, 1 chiffre, 1 majuscule et 1 caractère spécial"), AllowedMethods: nil})
				return
			}

			err = login.CreateAccount(db, username, email, password)
			if err != nil {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Erreur création du compte"), AllowedMethods: nil})
				return
			}

			userID, err := login.Login(db, email, password)
			if err != nil {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Email ou mot de passe incorrect"), AllowedMethods: nil})
				return
			}

			sessionID, expiresAt, err := m.CreationSession(userID)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			cookieUser := login.CreateCookie(sessionID, expiresAt, COOKIE_NAME)
			http.SetCookie(w, cookieUser)

			http.Redirect(w, r, "/cookies", http.StatusSeeOther)

		default:
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET", "POST"}})
			return
		}
	}
}

func Profile(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
			return
		}

		userSession, ok, err := m.IsSessionActive(r, COOKIE_NAME)
		if err != nil {
			renderError(w, 500, nil)
			return
		}
		if ok == false {
			renderError(w, 401, &ErrorDetails{Error: errors.New("Session invalide ou expirée"), AllowedMethods: nil})
			return
		}

		t, err := parseTemplates(PROFILE_PATH)
		if err != nil {
			renderError(w, 500, nil)
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
		JOIN roles ON users.roles_id = roles.id
		WHERE users.id = ?`, userSession.UserID)

		var user models.User
		var date string

		err = userSQL.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&date,
			&user.Role,
		)
		if err != nil {
			renderError(w, 500, nil)
			return
		}

		user.CreatedAt, err = time.Parse(time.DateTime, date)
		if err != nil {
			renderError(w, 500, nil)
			return
		}

		userData := session.Authentified{IsActive: true}
		PageData := models.PageData{Auth: userData, Profile: user}

		w, err = executeBuffer(t, w, PageData)
		if err != nil {
			renderError(w, 500, nil)
			return
		}
	}
}

func Cookies(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, COOKIES_PATH)
	} else {
		renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
		return
	}
}

// check pourquoi la déconnexion ne marche pas tout le temps
func Logout(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
			return
		}

		sessionID, err := login.GetSessionCookie(r, COOKIE_NAME)
		if err != nil {
			log.Println("Erreur dans la récup du nom de cookie :", err)
			renderError(w, 401, &ErrorDetails{Error: errors.New("Session invalide ou expirée"), AllowedMethods: nil})
			return
		}

		err = m.SuppressionSession(sessionID)
		if err != nil {
			log.Println("Erreur dans la suppression de la session :", err)
			renderError(w, 500, nil)
			return
		}
		login.DeleteSessionCookie(w, COOKIE_NAME)
		http.ServeFile(w, r, LOGOUT_PATH)
	}
}

// CreatePost create a post written by the user.
func CreatePost(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				log.Println(err)
				renderError(w, 500, nil)
				return
			}
			if ok == false {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Vous devez être connecté pour écrire un post"), AllowedMethods: nil})
				return
			}

			userData := session.Authentified{IsActive: true}
			pageData := models.PageData{Auth: userData}

			t, err := parseTemplates(CREATEPOST_PATH)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, pageData)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

		case http.MethodPost:

			userSession, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == false {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Session invalide ou expirée"), AllowedMethods: nil})
				return
			}

			userID := userSession.UserID

			title := strings.TrimSpace(r.FormValue("title"))
			content := strings.TrimSpace(r.FormValue("content"))
			categories := r.Form["categories"]

			if title == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le titre est obligatoire"), AllowedMethods: nil})
				return
			}

			if len(title) > 80 {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le titre est trop long. Veuillez mettre moins de 50 caractères"), AllowedMethods: nil})
				return
			}

			if content == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le contenu est obligatoire"), AllowedMethods: nil})
				return
			}

			if len(content) > 15000 {
				renderError(w, 400, &ErrorDetails{Error: errors.New("La taille du post a dépassé "), AllowedMethods: nil})
				return
			}

			categoryText := strings.Join(categories, ",")

			postID, err := uuid.NewV4()
			if err != nil {
				renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur création du post"), AllowedMethods: nil})
				return
			}

			_, err = db.Exec(`
			INSERT INTO posts (id, user_id, title, content, categories, created_at)
			VALUES (?, ?, ?, ?, ?, datetime('now'))
		`, postID.String(), userID, title, content, categoryText)

			if err != nil {
				renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur récupération des posts"), AllowedMethods: nil})
				return
			}
			http.Redirect(w, r, "/posts", http.StatusSeeOther)

		default:
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET", "POST"}})
			return
		}
	}
}

// ReadPosts affiche les publications du forum.
func ReadPosts(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
			return
		}

		t, err := parseTemplates(POSTS_PATH)
		if err != nil {
			renderError(w, 500, nil)
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
			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur récupération des posts"), AllowedMethods: nil})
			return
		}
		defer rows.Close()

		posts := make([]models.Post, 0)

		for rows.Next() {

			var post models.Post
			var categoryText string
			var date string

			err := rows.Scan(
				&post.ID,
				&post.UserID,
				&post.Username,
				&post.Title,
				&post.Content,
				&categoryText,
				&date,
			)
			if err != nil {
				renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur lecture des posts"), AllowedMethods: nil})
				return
			}

			post.CreatedAt, err = time.Parse(time.DateTime, date)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			if categoryText != "" {
				post.Categories = strings.Split(categoryText, ",")
			}
			posts = append(posts, post)
		}

		if err := rows.Err(); err != nil {
			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur lecture des posts"), AllowedMethods: nil})
			return
		}

		var userData session.Authentified

		_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
		if err != nil {
			renderError(w, 500, nil)
			return
		}
		if ok == true {
			userData.IsActive = true
		}
		pageData := models.PageData{Auth: userData, Posts: posts}

		w, err = executeBuffer(t, w, pageData)
		if err != nil {
			renderError(w, 500, nil)
			return
		}
	}
}

func ForgotPasswordHandler(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:

			var userData session.Authentified

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == true {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Vous êtes déjà connecté"), AllowedMethods: nil})
				return
			}

			pageData := models.PageData{Auth: userData}

			t, err := parseTemplates(FORGOT_PASSWORD)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, pageData)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
		case http.MethodPost:

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == true {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Vous êtes déjà connecté"), AllowedMethods: nil})
				return
			}

			email := r.FormValue("email")

			if email == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Email obligatoire"), AllowedMethods: nil})
				return
			}

			err = login.CreatePasswordResetCode(db, email)
			if err != nil {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Erreur dans l'envoi du code de reinitialisation du mot de passe"), AllowedMethods: nil})
				return
			}

			http.Redirect(w, r, "/reset-password", http.StatusSeeOther)

		default:
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET", "POST"}})
			return
		}
	}
}

func ResetPasswordHandler(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:

			var userData session.Authentified

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == true {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Veuillez vous déconnecter pour réinitialiser votre mot de passe"), AllowedMethods: nil})
				return
			}

			pageData := models.PageData{Auth: userData}

			t, err := parseTemplates(RESET_PASSWORD)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, pageData)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

		case http.MethodPost:

			_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == true {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Vous êtes déjà connecté"), AllowedMethods: nil})
				return
			}

			code := r.FormValue("code")
			password := r.FormValue("password")
			passwordConfirm := r.FormValue("password_confirm")

			if code == "" || password == "" || passwordConfirm == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Tous les champs sont obligatoires"), AllowedMethods: nil})
				return
			}

			if len(code) != 6 {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le code doit contenir 6 chiffres"), AllowedMethods: nil})
				return
			}

			for _, char := range code {
				if char < '0' || char > '9' {
					renderError(w, 400, &ErrorDetails{Error: errors.New("Le code doit contenir uniquement des chiffres"), AllowedMethods: nil})
					return
				}
			}

			if password != passwordConfirm {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Les mots de passe ne correspondent pas"), AllowedMethods: nil})
				return
			}

			userID, err := login.VerifyPasswordResetCode(db, code)
			if err != nil {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Code incorrect ou expiré"), AllowedMethods: nil})
				return
			}

			err = login.ResetPassword(db, userID, password)
			if err != nil {
				renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur modification du mot de passe"), AllowedMethods: nil})
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)

		default:
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET", "POST"}})
			return
		}
	}
}
