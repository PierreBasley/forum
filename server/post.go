package server

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"forum/login"

	"github.com/gofrs/uuid"
)

type Post struct {
	ID         string
	UserID     string
	Username   string
	Title      string
	Content    string
	Categories []string
	CreatedAt  time.Time
	Likes      int
	Dislikes   int
}

// CreatePost crée un post pour l'utilisateur connecté.
func CreatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

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
			INSERT INTO posts (
				id,
				user_id,
				title,
				content,
				categories,
				created_at
			)
			VALUES (?, ?, ?, ?, ?, ?)
		`,
			postID.String(),
			userID,
			title,
			content,
			categoryText,
			time.Now(),
		)

		if err != nil {
			log.Println("ERREUR SQL READ POSTS :", err)
			http.Error(w, "Erreur récupération des posts", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	}
}

// ReadPosts affiche les publications du forum.
func ReadPosts(db *sql.DB) http.HandlerFunc {
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
		posts.created_at,
		posts.likes,
		posts.dislikes
	FROM posts
	JOIN users ON users.id = posts.user_id
	ORDER BY posts.created_at DESC
`)

		if err != nil {
			http.Error(w, "Erreur récupération des posts", http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		posts := make([]Post, 0)

		for rows.Next() {

			var post Post
			var categoryText string

			err := rows.Scan(
				&post.ID,
				&post.UserID,
				&post.Username,
				&post.Title,
				&post.Content,
				&categoryText,
				&post.CreatedAt,
				&post.Likes,
				&post.Dislikes,
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
			return
		}

		tmpl, err := template.ParseFiles("templates/posts.html")
		if err != nil {
			http.Error(w, "Erreur chargement de la page", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, posts)
		if err != nil {
			http.Error(w, "Erreur affichage des publications", http.StatusInternalServerError)
			return
		}
	}
}
