package server

import (
	"database/sql"
	"fmt"
	"forum/login"
	"forum/models"
	"log"
	"net/http"
	"regexp"
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

func LoginHandler(db *sql.DB) http.HandlerFunc {
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

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hashDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hashUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hashSpecial := regexp.MustCompile(`[!@#~$%^&*()+|_.,<>?/{}\-]`).MatchString(password)

	return hashDigit && hashUpper && hashSpecial
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

			if !isValidPassword(password) {
				http.Error(
					w,
					"Le mot de passe doit contenir au minimum 8 caractères, 1 chiffre, 1 majuscule et 1 caractère spécial",
					http.StatusBadRequest,
				)
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

			/*code, err := login.GenerateTwoFactorCode()
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

			_, err = db.Exec(
				`DELETE FROM two_factor_codes WHERE user_id = ?`,
				userID,
			)

			if err != nil {
				http.Error(w, "Erreur préparation 2FA", http.StatusInternalServerError)
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
				http.Error(w, "Erreur enregistrement du code", http.StatusInternalServerError)
				return
			}

			err = login.SendTwoFactorEmail(email, code)
			if err != nil {
				fmt.Println("ERREUR SMTP :", err)

				db.Exec(
					`DELETE FROM two_factor_codes WHERE id = ?`,
					codeID.String(),
				)

				http.Error(
					w,
					"Impossible d'envoyer le code par email",
					http.StatusInternalServerError,
				)
				return
			}*/

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

func Profile(db *sql.DB) http.HandlerFunc {
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
			&user.Created_at,
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

func Logout(db *sql.DB) http.HandlerFunc {
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
func CreatePost(db *sql.DB) http.HandlerFunc {
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

// func CreatePostPage(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		if r.Method != http.MethodGet {
// 			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		t, err := parseTemplates(CREATEPOST_PATH)
// 		if err != nil {
// 			renderError(w, 500)
// 			return
// 		}

// 		sessionID, err := login.GetSessionCookie(r)
// 		if err != nil {
// 			http.Redirect(w, r, "/login", http.StatusSeeOther)
// 			return
// 		}

// 		var userID string

// 		err = db.QueryRow(`
// 			SELECT user_id
// 			FROM sessions
// 			WHERE id = ?
// 			AND expires_at > datetime('now')
// 		`, sessionID).Scan(&userID)

// 		if err != nil {
// 			http.Redirect(w, r, "/login", http.StatusSeeOther)
// 			return
// 		}

// 		w, err = executeBuffer(t, w)
// 		if err != nil {
// 			renderError(w, 500)
// 			return
// 		}
// 		//http.ServeFile(w, r, "templates/create-post.html")
// 	}
// }

// ReadPosts affiche les publications du forum.
func ReadPosts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		// Why is there likes and dislikes in this query ?
		// 		rows, err := db.Query(`
		// 	SELECT
		// 		posts.id,
		// 		posts.user_id,
		// 		users.username,
		// 		posts.title,
		// 		posts.content,
		// 		posts.categories,
		// 		posts.created_at,
		// 		posts.likes,
		// 		posts.dislikes
		// 	FROM posts
		// 	JOIN users ON users.id = posts.user_id
		// 	ORDER BY posts.created_at DESC
		// `)

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

			// err := rows.Scan(
			// 	&post.ID,
			// 	&post.UserID,
			// 	&post.Username,
			// 	&post.Title,
			// 	&post.Content,
			// 	&categoryText,
			// 	&post.CreatedAt,
			// 	&post.Likes,
			// 	&post.Dislikes,
			// )

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

		// tmpl, err := template.ParseFiles("templates/posts.html")
		// if err != nil {
		// 	http.Error(w, "Erreur chargement de la page", http.StatusInternalServerError)
		// 	return
		// }

		// err = tmpl.Execute(w, posts)
		// if err != nil {
		// 	http.Error(w, "Erreur affichage des publications", http.StatusInternalServerError)
		// 	return
		// }
	}
}

//Not useful anymore
// func LoginPage(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	http.ServeFile(w, r, "templates/index.html")
// }

/*
func TwoFactorPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	_, err := r.Cookie("pending_2fa")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_, err = r.Cookie("pending_user_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "templates/2fa.html")
}

func TwoFactorHandler(db *sql.DB) http.HandlerFunc {
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
		code := r.FormValue("code")

		blocked, err := login.IsTwoFactorBlocked(db, userID)
		if err != nil {
			http.Error(w, "Erreur vérification du compte", http.StatusInternalServerError)
			return
		}

		if blocked {
			http.Error(w, "Compte bloqué après 5 tentatives 2FA", http.StatusForbidden)
			return
		}

		if len(code) != 6 {
			http.Error(w, "Le code doit contenir 6 chiffres", http.StatusBadRequest)
			return
		}

		for _, char := range code {
			if char < '0' || char > '9' {
				http.Error(w, "Le code doit contenir uniquement des chiffres", http.StatusBadRequest)
				return
			}
		}

		codeHash := login.HashTwoFactorCode(code)

		var codeID string

		err = db.QueryRow(`
			SELECT id
			FROM two_factor_codes
			WHERE user_id = ?
			AND code_hash = ?
			AND expires_at > datetime('now')
			AND used = 0
			ORDER BY expires_at DESC
			LIMIT 1
		`, userID, codeHash).Scan(&codeID)

		if err != nil {
			attempts, attemptErr := login.AddTwoFactorAttempt(db, userID)

			if attemptErr != nil {
				http.Error(w, "Erreur gestion des tentatives", http.StatusInternalServerError)
				return
			}

			if attempts >= 5 {
				http.Error(w, "Compte bloqué après 5 tentatives 2FA", http.StatusForbidden)
				return
			}

			http.Error(
				w,
				fmt.Sprintf("Code incorrect ou expiré. Tentative %d/5.", attempts),
				http.StatusUnauthorized,
			)
			return
		}

		_, err = db.Exec(`
			UPDATE two_factor_codes
			SET used = 1
			WHERE id = ?
			AND used = 0
		`, codeID)

		if err != nil {
			http.Error(w, "Erreur validation du code", http.StatusInternalServerError)
			return
		}

		err = login.ResetTwoFactorAttempts(db, userID)
		if err != nil {
			http.Error(w, "Erreur réinitialisation des tentatives", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "pending_2fa",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		http.Redirect(w, r, "/cookies", http.StatusSeeOther)
	}
}

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "templates/forgot-password.html")
}

func ForgotPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")

		if email == "" {
			http.Error(w, "Email obligatoire", http.StatusBadRequest)
			return
		}

		err := login.CreatePasswordResetCode(db, email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/reset-password", http.StatusSeeOther)
	}
}

func ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "templates/reset-password.html")
}

func ResetPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		code := r.FormValue("code")
		password := r.FormValue("password")
		passwordConfirm := r.FormValue("password_confirm")

		if email == "" || code == "" || password == "" || passwordConfirm == "" {
			http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
			return
		}

		if len(code) != 6 {
			http.Error(w, "Le code doit contenir 6 chiffres", http.StatusBadRequest)
			return
		}

		for _, char := range code {
			if char < '0' || char > '9' {
				http.Error(w, "Le code doit contenir uniquement des chiffres", http.StatusBadRequest)
				return
			}
		}

		if password != passwordConfirm {
			http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
			return
		}

		userID, err := login.VerifyPasswordResetCode(db, email, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		err = login.ResetPassword(db, userID, password)
		if err != nil {
			http.Error(w, "Erreur modification du mot de passe", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

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
