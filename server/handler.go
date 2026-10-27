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
	"unicode/utf8"

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
	POST_PATH       = PATH + "post.html"
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

		// var userData session.Authentified

		// _, ok, err := m.IsSessionActive(r, COOKIE_NAME)
		// if err != nil {
		// 	renderError(w, 500, nil)
		// 	return
		// }
		// if ok == true {
		// 	userData.IsActive = true
		// }

		// t, err := parseTemplates(INDEX_PATH)
		// if err != nil {
		// 	renderError(w, 500, nil)
		// 	return
		// }

		// pageData := models.PageData{Auth: userData}
		// w, err = executeBuffer(t, w, pageData)
		// if err != nil {
		// 	renderError(w, 500, nil)
		// 	return
		// }

		http.Redirect(w, r, "/posts", http.StatusFound)
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
		w.Header().Set("Cache-Control", "no-store")
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

			if strings.TrimSpace(content) == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Le contenu est obligatoire"), AllowedMethods: nil})
				return
			}

			if utf8.RuneCountInString(content) > 15000 {
				renderError(w, 400, &ErrorDetails{Error: errors.New("La taille du post a dépassé la limite de caractères"), AllowedMethods: nil})
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

func ReadPosts(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"GET"}})
			return
		}

		_, ok, err := m.IsSessionActive(r, COOKIE_NAME)
		if err != nil {
			log.Println(err)
			renderError(w, 500, nil)
			return
		}

		userData := session.Authentified{IsActive: ok}

		var posts []models.Post

		rows, err := db.Query(`
		SELECT
			posts.id,
			posts.user_id,
			users.username,
			posts.title,
			posts.content,
			COALESCE(posts.categories, ''),
			posts.created_at,
			(
				SELECT COUNT(*)
				FROM likes
				WHERE likes.post_id = posts.id
				AND likes.reaction = 'like'
			) AS likes_count,
			(
				SELECT COUNT(*)
				FROM likes
				WHERE likes.post_id = posts.id
				AND likes.reaction = 'dislike'
			) AS dislikes_count,
			(
				SELECT COUNT(*)
				FROM comments
				WHERE comments.post_id = posts.id
			) AS replies_count
		FROM posts
		JOIN users ON users.id = posts.user_id
		ORDER BY posts.created_at DESC
		`)
		if err != nil {
			log.Println(err)
			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur récupération des posts"), AllowedMethods: nil})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var post models.Post
			var categoryText string
			var dateText string

			err := rows.Scan(
				&post.ID,
				&post.UserID,
				&post.Username,
				&post.Title,
				&post.Content,
				&categoryText,
				&dateText,
				&post.LikesCount,
				&post.DislikesCount,
				&post.RepliesCount,
			)
			if err != nil {
				log.Println("Erreur de lecture d'une publication :", err)
				renderError(w, 500, nil)
				return
			}

			post.CreatedAt, err = time.Parse(time.DateTime, dateText)
			if err != nil {
				log.Println("Erreur de conversion de la date :", err)
				renderError(w, 500, nil)
				return
			}

			if categoryText != "" {
				post.Categories = strings.Split(categoryText, ",")
			}

			post.Score = post.LikesCount - post.DislikesCount

			posts = append(posts, post)
		}
		err = rows.Err()
		if err != nil {
			log.Println("Erreur pendant le parcours des publications :", err)
			renderError(w, 500, nil)
			return
		}

		pageData := models.PostsPageData{
			Auth:  userData,
			Posts: posts,
		}

		t, err := parseTemplates(POSTS_PATH)
		if err != nil {
			log.Println("Erreur de chargement du template des publications :", err)
			renderError(w, 500, nil)
			return
		}

		_, err = executeBuffer(t, w, pageData)
		if err != nil {
			log.Println("Erreur de rendu ou d'envoi des publications :", err)
			renderError(w, 500, nil)
			return
		}
	}
}

func ReadPost(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			renderError(w, 405, &ErrorDetails{AllowedMethods: []string{"GET"}})
			return
		}

		postID := r.PathValue("post_id")

		parsedID, err := uuid.FromString(postID)
		if err != nil {
			renderError(w, 400, &ErrorDetails{Error: errors.New("Identifiant de publication invalide"), AllowedMethods: nil})
			return
		}

		var post models.Post
		var categoryText string
		var dateText string

		err = db.QueryRow(`
			SELECT
				posts.id,
				posts.user_id,
				users.username,
				posts.title,
				posts.content,
				COALESCE(posts.categories, ''),
				posts.created_at
			FROM posts
			JOIN users ON users.id = posts.user_id
			WHERE posts.id = ?
		`, parsedID.String()).Scan(
			&post.ID,
			&post.UserID,
			&post.Username,
			&post.Title,
			&post.Content,
			&categoryText,
			&dateText,
		)
		if errors.Is(err, sql.ErrNoRows) {
			renderError(w, 404, nil)
			return
		}
		if err != nil {
			log.Println("Erreur de récupération de la publication :", err)
			renderError(w, 500, nil)
			return
		}

		post.CreatedAt, err = time.Parse(time.DateTime, dateText)
		if err != nil {
			log.Println("Erreur de conversion de la date :", err)
			renderError(w, 500, nil)
			return
		}

		if categoryText != "" {
			post.Categories = strings.Split(categoryText, ",")
		}

		userData := session.Authentified{}
		userID := ""
		userSession, ok, err := m.IsSessionActive(r, COOKIE_NAME)
		if err != nil {
			log.Println("Erreur de vérification de la session :", err)
			renderError(w, 500, nil)
			return
		}
		if ok == true {
			userData = session.Authentified{IsActive: ok}
			userID = userSession.UserID
			userReact := ""

			err = db.QueryRow(`
			SELECT reaction
			FROM likes
			WHERE user_id = ? AND post_id = ?
			`, userID, post.ID).Scan(&userReact)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Println("Erreur dans la réaction de l'utilisateur:", err)
				renderError(w, 500, nil)
				return
			}
			if userReact == "like" {
				post.LikedByMe = true
				post.DislikedByMe = false
			}
			if userReact == "dislike" {
				post.DislikedByMe = true
				post.LikedByMe = false
			}
		}

		rows, err := db.Query(`
		SELECT
			comments.id,
			users.username,
			comments.content,
			comments.created_at,
			(
				SELECT COUNT(*)
				FROM comment_likes
				WHERE comment_likes.comment_id = comments.id
				AND comment_likes.reaction = 'like'
			) AS likes_count,
			(
				SELECT COUNT(*)
				FROM comment_likes
				WHERE comment_likes.comment_id = comments.id
				AND comment_likes.reaction = 'dislike'
			) AS dislikes_count,
			COALESCE(
				(
					SELECT reaction
					FROM comment_likes
					WHERE comment_likes.comment_id = comments.id
					AND comment_likes.user_id = ?
				),
				''
			) AS my_reaction
		FROM comments
		JOIN users ON users.id = comments.user_id
		WHERE comments.post_id = ?
		ORDER BY comments.created_at ASC, comments.id ASC
		`, userID, post.ID)
		if err != nil {
			log.Println("Erreur de récupération des commentaires :", err)
			renderError(w, 500, nil)
			return
		}
		defer rows.Close()

		var comments []models.Comment

		for rows.Next() {
			var comment models.Comment
			var commentDateText string
			var commentReaction string

			err := rows.Scan(
				&comment.ID,
				&comment.Username,
				&comment.Content,
				&commentDateText,
				&comment.LikesCount,
				&comment.DislikesCount,
				&commentReaction,
			)
			if err != nil {
				log.Println("Erreur de lecture d'un commentaire :", err)
				renderError(w, 500, nil)
				return
			}

			comment.CreatedAt, err = time.Parse(time.DateTime, commentDateText)
			if err != nil {
				log.Println("Erreur de conversion de la date d'un commentaire :", err)
				renderError(w, 500, nil)
				return
			}

			comment.LikedByMe = commentReaction == "like"
			comment.DislikedByMe = commentReaction == "dislike"
			comment.Score = comment.LikesCount - comment.DislikesCount
			comments = append(comments, comment)
		}

		err = rows.Err()
		if err != nil {
			log.Println("Erreur pendant le parcours des commentaires :", err)
			renderError(w, 500, nil)
			return
		}

		err = db.QueryRow(`
		SELECT COUNT(*)
			FROM likes
			WHERE post_id = ? and reaction = ?
		`, post.ID, "like").Scan(&post.LikesCount)
		if err != nil {
			log.Println("Erreur dans la lecture  dans le comptage de likes:", err)
			renderError(w, 500, nil)
			return
		}

		err = db.QueryRow(`
		SELECT COUNT(*)
			FROM likes
			WHERE post_id = ? and reaction = ?
		`, post.ID, "dislike").Scan(&post.DislikesCount)
		if err != nil {
			log.Println("Erreur dans la lecture  dans le comptage de dislikes:", err)
			renderError(w, 500, nil)
			return
		}
		post.Score = post.LikesCount - post.DislikesCount

		pageData := models.PostPageData{
			Auth:     userData,
			Post:     post,
			Comments: comments,
		}

		t, err := parseTemplates(POST_PATH)
		if err != nil {
			log.Println("Erreur de chargement du template de la publication :", err)
			renderError(w, 500, nil)
			return
		}

		_, err = executeBuffer(t, w, pageData)
		if err != nil {
			log.Println("Erreur de rendu ou d'envoi de la publication :", err)
			renderError(w, 500, nil)
			return
		}
	}
}

func Comment(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Checking that all needed data was provided
			if r.FormValue("post_id") == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Aucun ID de post n'a été fourni"), AllowedMethods: nil})
				return
			}
			// !!!: Right now, content isn't validated so major XSS flaws
			// might arise (even when filtered with Go templates)
			// In the future (and in a real-life project), this should be done
			// both input- and output-side (output is handled by templates)
			//PB : TODO seeing that with YF, putting a limit to the lenght of the comment
			content := r.FormValue("content")
			if strings.TrimSpace(content) == "" {
				renderError(w, 400, &ErrorDetails{Error: errors.New("Commentaire vide"), AllowedMethods: nil})
				return
			}

			if utf8.RuneCountInString(content) > 15000 {
				renderError(w, 400, &ErrorDetails{Error: errors.New("La taille de la réponse a dépassé la limite de caractères"), AllowedMethods: nil})
				return
			}
			// Check that the user is logged in
			//Maybe use the function isSessionActive instead ? it's refactoring
			sessionCookie, err := r.Cookie(COOKIE_NAME)
			if err != nil {
				if err == http.ErrNoCookie {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}
				log.Println(`Erreur de parsing du cookie de session pendant l'ajout d'un commentaire\n
				L'utilisateur n'a pas pu être authentifié`, err)
				renderError(w, 500, nil)
				return
			}
			userSession, validSession, err := m.RechercheSession(sessionCookie.Value)
			if err != nil {
				log.Println("Session invalide:", err)
				renderError(w, 500, nil)
				return
			}
			if !validSession {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			rows, err := db.Query(`
			SELECT id
			FROM users
			WHERE id = ?
			`, userSession.UserID)
			if err != nil {
				log.Println("erreur de requêtage sur le post pendant la création d'un commentaire:", err)
				renderError(w, 500, nil)
				return
			}
			rows.Next()
			returnedUserID := ""
			err = rows.Scan(&returnedUserID)
			if err != nil {
				log.Println("Session invalide contenant un utilisateur non-existant au moment de la requête, erreur de scan:", err.Error())
				renderError(w, 400, &ErrorDetails{Error: errors.New("session invalide"), AllowedMethods: nil})
				return
			}
			rows.Close()

			// Check that the post exists
			rows, err = db.Query(`
			SELECT id
			FROM posts
			WHERE id = ?
			`, r.FormValue("post_id"))
			if err != nil {
				log.Println("erreur de requêtage sur le post pendant la création d'un commentaire:", err)
				renderError(w, 500, nil)
				return
			}
			rows.Next()
			postID := ""
			if err = rows.Scan(&postID); err != nil {
				log.Println("erreur de requêtage sur le post pendant la création d'un commentaire, scanning de post_id:", err)
				renderError(w, 500, nil)
				return
			}
			rows.Close()

			// Insert the comment
			generated, err := uuid.NewV4()
			if err != nil {
				log.Println("erreur de requêtage sur le post pendant la création d'un commentaire, nouveau UUID:", err)
				renderError(w, 500, nil)
				return
			}
			newID := generated.String()
			_, err = db.Exec(`
			INSERT INTO comments (id, post_id, user_id, content, created_at)
			VALUES (?, ?, ?, ?, datetime('now'))
			`, newID, postID, returnedUserID, content, time.Now())
			if err != nil {
				log.Println("erreur d'insertion de commentaire,", err)
				renderError(w, 500, nil)
				return
			}
			rows.Close()
			http.Redirect(w, r, "/posts/"+postID, http.StatusFound)
		} else if r.Method == http.MethodPut {
			// this is where we handle the comment modifying logic
		} else {
			renderError(w, 405, &ErrorDetails{Error: nil, AllowedMethods: []string{"POST", "PUT"}})
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

// LikeHandler add, move out or replace the post reaction.
func LikeHandler(db *sql.DB, m *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			renderError(w, http.StatusMethodNotAllowed, &ErrorDetails{
				AllowedMethods: []string{"POST"},
			})
			return
		}

		// Vérifier que l'utilisateur est connecté.
		sessionID, err := login.GetSessionCookie(r, COOKIE_NAME)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userSession, ok, err := m.RechercheSession(sessionID)
		if err != nil {
			log.Println("Erreur récupération session :", err)
			renderError(w, 500, nil)
			return
		}
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Récupérer les données du formulaire.
		if err := r.ParseForm(); err != nil {
			renderError(w, 400, &ErrorDetails{Error: errors.New("Formulaire invalide"), AllowedMethods: nil})
			return
		}

		targetType := strings.TrimSpace(r.FormValue("target_type"))
		targetID := strings.TrimSpace(r.FormValue("target_id"))
		reaction := strings.TrimSpace(r.FormValue("reaction"))

		parsedID, err := uuid.FromString(targetID)
		if err != nil {
			renderError(w, 400, &ErrorDetails{Error: errors.New("Identifiant de cible invalide"), AllowedMethods: nil})
			return
		}

		targetID = parsedID.String()

		// Accepter uniquement les deux réactions autorisées.
		if reaction != "like" && reaction != "dislike" {
			renderError(w, 400, &ErrorDetails{Error: errors.New("Réaction invalide"), AllowedMethods: nil})
			return
		}

		userID := userSession.UserID

		targetQuery := ""
		reactionQuery := ""
		deleteQuery := ""
		updateQuery := ""
		insertQuery := ""

		switch targetType {
		case "post":

			targetQuery = "SELECT id FROM posts WHERE id = ?"
			reactionQuery = "SELECT reaction FROM likes WHERE user_id = ? AND post_id = ?"
			deleteQuery = "DELETE FROM likes WHERE user_id = ? AND post_id = ?"
			updateQuery = "UPDATE likes SET reaction = ? WHERE user_id = ? AND post_id = ?"
			insertQuery = "INSERT INTO likes (user_id, post_id, reaction) VALUES (?, ?, ?)"

		case "comment":
			targetQuery = "SELECT post_id FROM comments WHERE id = ?"
			reactionQuery = "SELECT reaction FROM comment_likes WHERE user_id = ? AND comment_id = ?"
			deleteQuery = "DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?"
			updateQuery = "UPDATE comment_likes SET reaction = ? WHERE user_id = ? AND comment_id = ?"
			insertQuery = "INSERT INTO comment_likes (user_id, comment_id, reaction) VALUES (?, ?, ?)"

		default:
			renderError(w, 400, &ErrorDetails{Error: errors.New("Erreur : le type n'est pas un post ou un comment"), AllowedMethods: nil})
			return
		}

		// Commencer une transaction.
		tx, err := db.Begin()
		if err != nil {
			log.Println("Erreur début transaction :", err)
			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur serveur"), AllowedMethods: nil})
			return
		}
		defer tx.Rollback()

		//checking for a existing ID
		var postID string

		err = tx.QueryRow(targetQuery, targetID).Scan(&postID)
		if errors.Is(err, sql.ErrNoRows) {
			renderError(w, 404, &ErrorDetails{Error: errors.New("La publication ou la réponse demandée n'existe pas"), AllowedMethods: nil})
			return
		}
		if err != nil {
			log.Println("Erreur de vérification de la cible :", err)
			renderError(w, 500, nil)
			return
		}

		// Chercher la réaction existante de cet utilisateur.
		var existingReaction string

		err = tx.QueryRow(reactionQuery, userID, targetID).Scan(&existingReaction)

		switch {
		case err == nil && existingReaction == reaction:
			_, err = tx.Exec(deleteQuery, userID, targetID)

		case err == nil:
			_, err = tx.Exec(updateQuery, reaction, userID, targetID)

		case errors.Is(err, sql.ErrNoRows):
			_, err = tx.Exec(insertQuery, userID, targetID, reaction)
		}
		if err != nil {
			log.Println("Erreur modification réaction :", err)
			renderError(w, 500, &ErrorDetails{Error: errors.New("Impossible de modifier la réaction"), AllowedMethods: nil})
			return
		}

		// Valider la transaction.
		if err := tx.Commit(); err != nil {
			log.Println("Erreur validation transaction :", err)
			renderError(w, 500, &ErrorDetails{Error: errors.New("Erreur serveur"), AllowedMethods: nil})
			return
		}

		// Revenir à la publication concernée
		http.Redirect(w, r, "/posts/"+postID, http.StatusSeeOther)
	}
}
