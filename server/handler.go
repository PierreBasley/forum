package server

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"forum/login"
	"math/big"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/gofrs/uuid"
)

func Index(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func Visitor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	http.Redirect(w, r, "/posts", http.StatusSeeOther)
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

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if username == "" || email == "" || password == "" {
			http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
			return
		}

		err := login.CreateAccount(db, username, email, password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID, err := login.Login(db, email, password)
		if err != nil {
			http.Error(w, "Erreur récupération du compte", http.StatusInternalServerError)
			return
		}

		code, err := generateTwoFactorCode()
		if err != nil {
			http.Error(w, "Erreur génération du code", http.StatusInternalServerError)
			return
		}

		codeHash := hashTwoFactorCode(code)

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

		err = sendTwoFactorEmail(email, code)
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
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "pending_user_id",
			Value:    userID,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "pending_2fa",
			Value:    "1",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		http.Redirect(w, r, "/2fa", http.StatusSeeOther)
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

func CreatePostPage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		sessionID, err := login.GetSessionCookie(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
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
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		http.ServeFile(w, r, "templates/create-post.html")
	}
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "templates/index.html")
}

func generateTwoFactorCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashTwoFactorCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

func sendTwoFactorEmail(destination string, code string) error {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpEmail == "" || smtpPassword == "" {
		fmt.Println("DEBUG SMTP_EMAIL présent :", smtpEmail != "")
		fmt.Println("DEBUG SMTP_PASSWORD présent :", smtpPassword != "")
		return fmt.Errorf("SMTP_EMAIL ou SMTP_PASSWORD manquant")
	}

	auth := smtp.PlainAuth(
		"",
		smtpEmail,
		smtpPassword,
		"smtp.gmail.com",
	)

	message := []byte(
		"From: " + smtpEmail + "\r\n" +
			"To: " + destination + "\r\n" +
			"Subject: Code de verification - Forum\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			"Votre code de verification est : " + code + "\r\n\r\n" +
			"Ce code est valable pendant 5 minutes.\r\n" +
			"Si vous n'etes pas a l'origine de cette connexion, ignorez cet email.\r\n",
	)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		smtpEmail,
		[]string{destination},
		message,
	)
}

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

		codeHash := hashTwoFactorCode(code)

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
			http.Error(w, "Code incorrect ou expiré", http.StatusUnauthorized)
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
