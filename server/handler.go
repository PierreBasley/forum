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

		t, err := parseTemplates(INDEX_PATH)
		if err != nil {
			renderError(w, 500, nil)
			return
		}

		w, err = executeBuffer(t, w, nil)
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
			t, err := parseTemplates(LOGIN_PATH)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, nil)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

		case http.MethodPost:

			email := r.FormValue("email")
			password := r.FormValue("password")

			userID, err := login.Login(db, email, password)
			if err != nil {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Email ou mot de passe incorrect"), AllowedMethods: nil})
				//http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
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
			t, err := parseTemplates(REGISTER_PATH)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, nil)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

		case http.MethodPost:

			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")
			passwordConfirm := r.FormValue("password_confirm")

			if username == "" || email == "" || password == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Tous les champs sont obligatoires"), AllowedMethods: nil})
				//http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
				return
			}

			if password != passwordConfirm {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Les mots de passe ne correspondent pas"), AllowedMethods: nil})
				//http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
				return
			}

			if login.IsValidPassword(password) == false {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le mot de passe doit contenir au minimum 8 caractères, 1 chiffre, 1 majuscule et 1 caractère spécial"), AllowedMethods: nil})
				//http.Error(w, "Le mot de passe doit contenir au minimum 8 caractères, 1 chiffre, 1 majuscule et 1 caractère spécial", http.StatusBadRequest)
				return
			}

			err := login.CreateAccount(db, username, email, password)
			if err != nil {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Erreur création du compte"), AllowedMethods: nil})
				//TODO : it shouldn't be an error 500 ?
				//http.Error(w, "Erreur création du compte", http.StatusBadRequest)
				return
			}

			userID, err := login.Login(db, email, password)
			if err != nil {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Email ou mot de passe incorrect"), AllowedMethods: nil})
				//http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
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

		t, err := parseTemplates(PROFILE_PATH)
		if err != nil {
			renderError(w, 500, nil)
			return
		}

		sessionID, err := login.GetSessionCookie(r, COOKIE_NAME)
		if err != nil {
			renderError(w, 401, &ErrorDetails{Error: errors.New("Session invalide ou expirée"), AllowedMethods: nil})
			//http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
			return
		}

		session, ok, err := m.RechercheSession(sessionID)
		if err != nil {
			renderError(w, 500, nil)
			return
		}
		if ok == false {
			renderError(w, 401, &ErrorDetails{Error: errors.New("Session invalide ou expirée"), AllowedMethods: nil})
			//http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
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
		WHERE users.id = ?`, session.UserID)

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
			//Todo : maybe remove the log later
			log.Println("problème dans le chargement des données de l'utilisateur:", err)
			renderError(w, 500, nil)
			return
		}
		user.CreatedAt, err = time.Parse(time.DateTime, date)
		if err != nil {
			//Todo : maybe remove the log later
			log.Println("problème dans la conversion de la date", err)
			renderError(w, 500, nil)
			return
		}
		w, err = executeBuffer(t, w, user)
		if err != nil {
			//Todo : maybe remove the log later
			log.Println("problème dans le rendu de la page profil", err)
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

func Logout(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
			return
		}

		sessionID, err := login.GetSessionCookie(r, COOKIE_NAME)
		if err != nil {
			renderError(w, 401, &ErrorDetails{Error: errors.New("Session invalide ou expirée"), AllowedMethods: nil})
			//http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
			return
		}

		err = m.SuppressionSession(sessionID)
		if err != nil {
			//Todo : maybe remove the log later
			log.Println("Erreur lors de la suppression de la session :", err)
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
			t, err := parseTemplates(CREATEPOST_PATH)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, nil)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

		case http.MethodPost:
			sessionID, err := login.GetSessionCookie(r, COOKIE_NAME)
			if err != nil {
				log.Println("Session ID problème get session", err)
				renderError(w, 401, &ErrorDetails{Error: errors.New("Session invalide ou expirée"), AllowedMethods: nil})
				//http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
				return
			}
			userSession, ok, err := m.RechercheSession(sessionID)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
			if ok == false {
				log.Println("Session introuvable ou expirée")
				renderError(w, 401, &ErrorDetails{Error: errors.New("Session invalide ou expirée"), AllowedMethods: nil})
				//http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
				return
			}

			userID := userSession.UserID

			title := strings.TrimSpace(r.FormValue("title"))
			content := strings.TrimSpace(r.FormValue("content"))
			categories := r.Form["categories"]

			if title == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le titre est obligatoire"), AllowedMethods: nil})
				//http.Error(w, "Le titre est obligatoire", http.StatusBadRequest)
				return
			}

			if content == "" {
				renderError(w, 400, nil)
				return
			}

			categoryText := strings.Join(categories, ",")

			postID, err := uuid.NewV4()
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			_, err = db.Exec(`
			INSERT INTO posts (id, user_id, title, content, categories, created_at)
			VALUES (?, ?, ?, ?, ?, datetime('now'))
		`, postID.String(), userID, title, content, categoryText)

			if err != nil {
				//Todo : maybe remove the log later
				log.Println("ERREUR SQL READ POSTS :", err)
				renderError(w, 500, nil)
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
			//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		t, err := parseTemplates(POSTS_PATH)
		if err != nil {
			log.Printf("erreur de template; %s", err.Error())
			renderError(w, 500, nil)
			return
		}

		var posts models.PostsData
		if r.PathValue("post_id") != "" {
			// Fetching post
			rows, err := db.Query(`
			SELECT
				posts.id,
				posts.user_id,
				users.username,
				title,
				content,
				categories,
				posts.created_at
			FROM posts
			JOIN users ON users.id = posts.user_id
			WHERE posts.id = ?
			`, r.PathValue("post_id"))
			if err != nil {
				log.Printf("erreur de requêtage sur posts, post_id = %s; %s", r.PathValue("post_id"), err.Error())
				renderError(w, 500, nil)
				return
			}
			post := models.Post{}
			date := ""
			if !rows.Next() {
				log.Println("erreur dans la requête (single post)", rows.Err())
				renderError(w, 500, nil)
				return
			}
			categories := ""
			err = rows.Scan(
				&post.ID, &post.UserID, &post.Username, &post.Title,
				&post.Content, &categories, &date,
			)
			if err != nil {
				log.Printf("erreur de scan sur single post, post_id = %s; %s", r.PathValue("post_id"), err.Error())
				renderError(w, 500, nil)
				return
			}
			post.Categories = strings.Split(categories, ",")
			post.CreatedAt, err = time.Parse(time.DateTime, date)
			if err != nil {
				log.Printf("erreur de parsing de la date sur single post, post_id = %s; %s", r.PathValue("post_id"), err.Error())
				renderError(w, 500, nil)
				return
			}
			rows.Close()
			posts.Posts = []models.Post{post}

			// Fetching comments
			rows, err = db.Query(`
			SELECT
				users.username,
				content,
				comments.created_at
			FROM comments
			JOIN users ON users.id = comments.user_id
			WHERE comments.post_id = ?
			ORDER BY comments.created_at DESC
			`, r.PathValue("post_id"))
			if err != nil {
				log.Printf("erreur de requêtage sur comments, post_id = %s; %s", r.PathValue("post_id"), err.Error())
				renderError(w, 500, nil)
				return
			}

			var comments []models.Comment = nil
			for rows.Next() {
				date := ""
				c := models.Comment{}
				err = rows.Scan(&c.Username, &c.Content, &date)
				if err != nil {
					log.Printf("erreur de scan sur commentaire, post_id = %s; %s", r.PathValue("post_id"), err.Error())
					renderError(w, 500, nil)
					return
				}
				c.CreatedAt, err = time.Parse(time.DateTime, date)
				if err != nil {
					log.Println("problème de conversion de date (commentaire):", err)
					renderError(w, 500, nil)
					return
				}
				comments = append(comments, c)
			}
			log.Println(comments)
			posts.Comments = comments
			rows.Close()
		} else {
			rows, err := db.Query(`
			SELECT
				posts.id,
				user_id,
				users.username,
				title,
				content,
				categories,
				posts.created_at
			FROM posts
			JOIN users ON users.id = posts.user_id
			ORDER BY posts.created_at DESC
			`)
			if err != nil {
				log.Printf("erreur de requêtage sur posts, post_id = %s; %s", r.PathValue("post_id"), err.Error())
				renderError(w, 500, nil)
				return
			}
			defer rows.Close()

			posts.Posts = make([]models.Post, 0)
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
					log.Printf("erreur de scan sur post, post_id = %s; %s", r.PathValue("post_id"), err.Error())
					renderError(w, 500, nil)
					return
				}

				post.CreatedAt, err = time.Parse(time.DateTime, date)
				if err != nil {
					log.Println("problème dans la conversion de la date", err)
					renderError(w, 500, nil)
					return
				}

				if categoryText != "" {
					post.Categories = strings.Split(categoryText, ",")
				}
				posts.Posts = append(posts.Posts, post)
			}

			if err := rows.Err(); err != nil {
				log.Println(err)
				renderError(w, 500, nil)
				return
			}
		}

		w, err = executeBuffer(t, w, posts)
		if err != nil {
			log.Println("ReadPosts : erreur de rendu dans l'execution du buffer", err)
			renderError(w, 500, nil)
			return
		}
	}
}

func ForgotPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			//http.ServeFile(w, r, "templates/pages/forgot-password.html")
			t, err := parseTemplates(FORGOT_PASSWORD)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, nil)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
		case http.MethodPost:
			email := r.FormValue("email")

			if email == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Email obligatoire"), AllowedMethods: nil})
				//http.Error(w, "Email obligatoire", http.StatusBadRequest)
				return
			}

			err := login.CreatePasswordResetCode(db, email)
			if err != nil {
				log.Println(err)
				renderError(w, 400, &ErrorDetails{Error: errors.New("Erreur dans l'envoi du code de reinitialisation du mot de passe"), AllowedMethods: nil})
				//http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			http.Redirect(w, r, "/reset-password", http.StatusSeeOther)

		default:
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET", "POST"}})
			//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
	}
}

func ResetPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			//http.ServeFile(w, r, "templates/pages/reset-password.html")
			t, err := parseTemplates(RESET_PASSWORD)
			if err != nil {
				renderError(w, 500, nil)
				return
			}

			w, err = executeBuffer(t, w, nil)
			if err != nil {
				renderError(w, 500, nil)
				return
			}
		case http.MethodPost:
			code := r.FormValue("code")
			password := r.FormValue("password")
			passwordConfirm := r.FormValue("password_confirm")

			if code == "" || password == "" || passwordConfirm == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Tous les champs sont obligatoires"), AllowedMethods: nil})
				//http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
				return
			}

			if len(code) != 6 {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le code doit contenir 6 chiffres"), AllowedMethods: nil})
				//http.Error(w, "Le code doit contenir 6 chiffres", http.StatusBadRequest)
				return
			}

			for _, char := range code {
				if char < '0' || char > '9' {
					renderError(w, 400, &ErrorDetails{Error: errors.New("Le code doit contenir uniquement des chiffres"), AllowedMethods: nil})
					//http.Error(w, "Le code doit contenir uniquement des chiffres", http.StatusBadRequest)
					return
				}
			}

			if password != passwordConfirm {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Les mots de passe ne correspondent pas"), AllowedMethods: nil})
				//http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
				return
			}

			userID, err := login.VerifyPasswordResetCode(db, code)
			if err != nil {
				renderError(w, 401, &ErrorDetails{Error: errors.New("Code incorrect ou expiré"), AllowedMethods: nil})
				//http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			err = login.ResetPassword(db, userID, password)
			if err != nil {
				renderError(w, 500, nil)
				//http.Error(w, "Erreur modification du mot de passe", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)

		default:
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET", "POST"}})
			//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
	}
}

// TODO : check that function
// func TwoFactorPage(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
// 		//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	_, err := r.Cookie("pending_2fa")
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusSeeOther)
// 		return
// 	}

// 	_, err = r.Cookie("pending_user_id")
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusSeeOther)
// 		return
// 	}

// 	http.ServeFile(w, r, "templates/2fa.html")
// }

// TODO : check that function
// func TwoFactorHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"POST"}})
// 			//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		cookie, err := r.Cookie("pending_user_id")
// 		if err != nil {
// 			renderError(w, 401, &ErrorDetails{Error: errors.New("Aucune authentification en attente"), AllowedMethods: nil})
// 			//http.Error(w, "Aucune authentification en attente", http.StatusUnauthorized)
// 			return
// 		}

// 		userID := cookie.Value
// 		code := r.FormValue("code")

// 		blocked, err := login.IsTwoFactorBlocked(db, userID)
// 		if err != nil {
// 			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur vérification du compte"), AllowedMethods: nil})
// 			//http.Error(w, "Erreur vérification du compte", http.StatusInternalServerError)
// 			return
// 		}

// 		if blocked {
// 			renderError(w, 403, &ErrorDetails{Error: errors.New("Compte bloqué après 5 tentatives 2FA"), AllowedMethods: nil})
// 			//http.Error(w, "Compte bloqué après 5 tentatives 2FA", http.StatusForbidden)
// 			return
// 		}

// 		if len(code) != 6 {
// 			renderError(w, 400, &ErrorDetails{Error: errors.New("Le code doit contenir 6 chiffres"), AllowedMethods: nil})
// 			//http.Error(w, "Le code doit contenir 6 chiffres", http.StatusBadRequest)
// 			return
// 		}

// 		for _, char := range code {
// 			if char < '0' || char > '9' {
// 				renderError(w, 400, &ErrorDetails{Error: errors.New("Le code doit contenir uniquement des chiffres"), AllowedMethods: nil})
// 				//http.Error(w, "Le code doit contenir uniquement des chiffres", http.StatusBadRequest)
// 				return
// 			}
// 		}

// 		codeHash := login.HashTwoFactorCode(code)

// 		var codeID string

// 		err = db.QueryRow(`
// 			SELECT id
// 			FROM two_factor_codes
// 			WHERE user_id = ?
// 			AND code_hash = ?
// 			AND expires_at > datetime('now')
// 			AND used = 0
// 			ORDER BY expires_at DESC
// 			LIMIT 1
// 		`, userID, codeHash).Scan(&codeID)

// 		if err != nil {
// 			attempts, attemptErr := login.AddTwoFactorAttempt(db, userID)

// 			if attemptErr != nil {
// 				renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur gestion des tentatives"), AllowedMethods: nil})
// 				//http.Error(w, "Erreur gestion des tentatives", http.StatusInternalServerError)
// 				return
// 			}

// 			if attempts >= 5 {
// 				renderError(w, 403, &ErrorDetails{Error: errors.New("Compte bloqué après 5 tentatives 2FA"), AllowedMethods: nil})
// 				//http.Error(w, "Compte bloqué après 5 tentatives 2FA", http.StatusForbidden)
// 				return
// 			}
// 			renderError(w, 401, &ErrorDetails{Error: errors.New("Code incorrect ou expiré. Tentative" + strconv.Itoa(attempts) + "/5"), AllowedMethods: nil})
// 			//http.Error(w,fmt.Sprintf("Code incorrect ou expiré. Tentative %d/5.", attempts),http.StatusUnauthorized)
// 			return
// 		}

// 		_, err = db.Exec(`
// 			UPDATE two_factor_codes
// 			SET used = 1
// 			WHERE id = ?
// 			AND used = 0
// 		`, codeID)

// 		if err != nil {
// 			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur validation du code"), AllowedMethods: nil})
// 			//http.Error(w, "Erreur validation du code", http.StatusInternalServerError)
// 			return
// 		}

// 		err = login.ResetTwoFactorAttempts(db, userID)
// 		if err != nil {
// 			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur réinitialisation des tentatives"), AllowedMethods: nil})
// 			//http.Error(w, "Erreur réinitialisation des tentatives", http.StatusInternalServerError)
// 			return
// 		}

// 		//TODO ; why setting a cookie ?
// 		http.SetCookie(w, &http.Cookie{
// 			Name:     "pending_2fa",
// 			Value:    "",
// 			Expires:  time.Unix(0, 0),
// 			HttpOnly: true,
// 			Secure:   false,
// 			SameSite: http.SameSiteStrictMode,
// 			Path:     "/",
// 		})

// 		http.Redirect(w, r, "/cookies", http.StatusSeeOther)
// 	}
// }

// func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
// 		//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	//TODO : putting the URL into a constant
// 	//http.ServeFile(w, r, "templates/pages/forgot-password.html")
// 	t, err := parseTemplates(FORGOT_PASSWORD)
// 	if err != nil {
// 		renderError(w, 500, nil)
// 		return
// 	}

// 	w, err = executeBuffer(t, w, nil)
// 	if err != nil {
// 		renderError(w, 500, nil)
// 		return
// 	}
// }

// func ForgotPasswordHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"POST"}})
// 			//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		email := r.FormValue("email")

// 		if email == "" {
// 			renderError(w, 400, &ErrorDetails{Error: errors.New("Email obligatoire"), AllowedMethods: nil})
// 			//http.Error(w, "Email obligatoire", http.StatusBadRequest)
// 			return
// 		}

// 		err := login.CreatePasswordResetCode(db, email)
// 		if err != nil {
// 			renderError(w, 400, &ErrorDetails{Error: errors.New("Erreur dans l'envoi du code de reinitialisation du mot de passe"), AllowedMethods: nil})
// 			//http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}

// 		http.Redirect(w, r, "/reset-password", http.StatusSeeOther)
// 	}
// }

// func ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
// 		//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	//http.ServeFile(w, r, "templates/pages/reset-password.html")
// 	t, err := parseTemplates("templates/pages/reset-password.html")
// 	if err != nil {
// 		renderError(w, 500, nil)
// 		return
// 	}

// 	w, err = executeBuffer(t, w, nil)
// 	if err != nil {
// 		renderError(w, 500, nil)
// 		return
// 	}

// }

// func ResetPasswordHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"POST"}})
// 			//http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		code := r.FormValue("code")
// 		password := r.FormValue("password")
// 		passwordConfirm := r.FormValue("password_confirm")

// 		if code == "" || password == "" || passwordConfirm == "" {
// 			renderError(w, 400, &ErrorDetails{Error: errors.New("Tous les champs sont obligatoires"), AllowedMethods: nil})
// 			//http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
// 			return
// 		}

// 		if len(code) != 6 {
// 			renderError(w, 400, &ErrorDetails{Error: errors.New("Le code doit contenir 6 chiffres"), AllowedMethods: nil})
// 			//http.Error(w, "Le code doit contenir 6 chiffres", http.StatusBadRequest)
// 			return
// 		}

// 		for _, char := range code {
// 			if char < '0' || char > '9' {
// 				renderError(w, 400, &ErrorDetails{Error: errors.New("Le code doit contenir uniquement des chiffres"), AllowedMethods: nil})
// 				//http.Error(w, "Le code doit contenir uniquement des chiffres", http.StatusBadRequest)
// 				return
// 			}
// 		}

// 		if password != passwordConfirm {
// 			renderError(w, 400, &ErrorDetails{Error: errors.New("Les mots de passe ne correspondent pas"), AllowedMethods: nil})
// 			//http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
// 			return
// 		}

// 		userID, err := login.VerifyPasswordResetCode(db, code)
// 		if err != nil {
// 			renderError(w, 401, &ErrorDetails{Error: errors.New("Code incorrect ou expiré"), AllowedMethods: nil})
// 			//http.Error(w, err.Error(), http.StatusUnauthorized)
// 			return
// 		}

// 		err = login.ResetPassword(db, userID, password)
// 		if err != nil {
// 			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur modification du mot de passe"), AllowedMethods: nil})
// 			http.Error(w, "Erreur modification du mot de passe", http.StatusInternalServerError)
// 			return
// 		}

// 		http.Redirect(w, r, "/login", http.StatusSeeOther)
// 	}
// }

/*
func ResendTwoFactorHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		cookie, err := r.Cookie("pending_user_id")
		if err != nil {
			http.Error(w, "Aucune authentification en attente", http.StatusUnauthorized)
			return
		}

		userID := cookie.Value

		blocked, err := login.IsTwoFactorBlocked(db, userID)
		if err != nil {
			http.Error(w, "Erreur vérification du compte", http.StatusInternalServerError)
			return
		}

		if blocked {
			http.Error(w, "Compte bloqué après 5 tentatives 2FA", http.StatusForbidden)
			return
		}

		var email string

		err = db.QueryRow(`
			SELECT email
			FROM users
			WHERE id = ?
		`, userID).Scan(&email)

		if err != nil {
			http.Error(w, "Utilisateur introuvable", http.StatusNotFound)
			return
		}

		code, err := login.GenerateTwoFactorCode()
		if err != nil {
			http.Error(w, "Erreur génération du code", http.StatusInternalServerError)
			return
		}

		codeHash := login.HashTwoFactorCode(code)

		codeID, err := uuid.NewV4()
		if err != nil {
			http.Error(w, "Erreur génération du code", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
			DELETE FROM two_factor_codes
			WHERE user_id = ?
		`, userID)

		if err != nil {
			http.Error(w, "Erreur préparation du nouveau code", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`
			INSERT INTO two_factor_codes (
				id,
				user_id,
				code_hash,
				expires_at,
				used
			)
			VALUES (?, ?, ?, datetime('now', '+5 minutes'), 0)
		`, codeID.String(), userID, codeHash)

		if err != nil {
			http.Error(w, "Erreur enregistrement du nouveau code", http.StatusInternalServerError)
			return
		}

		err = login.SendTwoFactorEmail(email, code)
		if err != nil {
			_, _ = db.Exec(
				`DELETE FROM two_factor_codes WHERE id = ?`,
				codeID.String(),
			)

			http.Error(
				w,
				"Impossible d'envoyer le nouveau code par email",
				http.StatusInternalServerError,
			)
			return
		}

		http.Redirect(w, r, "/2fa", http.StatusSeeOther)
	}

}
*/
