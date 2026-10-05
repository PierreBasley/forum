package server

// type Post struct {
// 	ID         string
// 	UserID     string
// 	Username   string
// 	Title      string
// 	Content    string
// 	Categories []string
// 	CreatedAt  time.Time
// 	Likes      int
// 	Dislikes   int
// }

// // CreatePost crée un post pour l'utilisateur connecté.
// func CreatePost(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		if r.Method != http.MethodPost {
// 			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		sessionID, err := login.GetSessionCookie(r)
// 		if err != nil {
// 			http.Error(w, "Vous devez être connecté", http.StatusUnauthorized)
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
// 			http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
// 			return
// 		}

// 		title := strings.TrimSpace(r.FormValue("title"))
// 		content := strings.TrimSpace(r.FormValue("content"))
// 		categories := r.Form["categories"]

// 		if title == "" {
// 			http.Error(w, "Le titre est obligatoire", http.StatusBadRequest)
// 			return
// 		}

// 		if content == "" {
// 			http.Error(w, "Le contenu est obligatoire", http.StatusBadRequest)
// 			return
// 		}

// 		categoryText := strings.Join(categories, ",")

// 		postID, err := uuid.NewV4()
// 		if err != nil {
// 			http.Error(w, "Erreur création du post", http.StatusInternalServerError)
// 			return
// 		}

// 		_, err = db.Exec(`
// 			INSERT INTO posts (
// 				id,
// 				user_id,
// 				title,
// 				content,
// 				categories,
// 				created_at
// 			)
// 			VALUES (?, ?, ?, ?, ?, ?)
// 		`,
// 			postID.String(),
// 			userID,
// 			title,
// 			content,
// 			categoryText,
// 			time.Now(),
// 		)

// 		if err != nil {
// 			log.Println("ERREUR SQL READ POSTS :", err)
// 			http.Error(w, "Erreur récupération des posts", http.StatusInternalServerError)
// 			return
// 		}

// 		http.Redirect(w, r, "/posts", http.StatusSeeOther)
// 	}
// }
