package login

//
//import (
//	"database/sql"
//	"errors"
//	"fmt"
//	"net/smtp"
//	"os"
//	"time"
//
//	"github.com/gofrs/uuid"
//	"golang.org/x/crypto/bcrypt"
//)
//
//func CreatePasswordResetCode(db *sql.DB, email string) error {
//	var userID string
//
//	err := db.QueryRow(`
//		SELECT id
//		FROM users
//		WHERE email = ?
//	`, email).Scan(&userID)
//
//	if err != nil {
//		return errors.New("email introuvable")
//	}
//
//	code, err := GenerateTwoFactorCode()
//	if err != nil {
//		return err
//	}
//
//	codeHash := HashTwoFactorCode(code)
//
//	codeID, err := uuid.NewV4()
//	if err != nil {
//		return err
//	}
//
//	_, err = db.Exec(`
//		DELETE FROM password_reset_codes
//		WHERE user_id = ?
//	`, userID)
//
//	if err != nil {
//		return err
//	}
//
//	_, err = db.Exec(`
//		INSERT INTO password_reset_codes (
//			id,
//			user_id,
//			code_hash,
//			expires_at,
//			used
//		)
//		VALUES (?, ?, ?, datetime('now', '+5 minutes'), 0)
//	`, codeID.String(), userID, codeHash)
//
//	if err != nil {
//		return err
//	}
//
//	if err := SendPasswordResetEmail(email, code); err != nil {
//		_, _ = db.Exec(
//			`DELETE FROM password_reset_codes WHERE id = ?`,
//			codeID.String(),
//		)
//
//		return err
//	}
//
//	return nil
//}
//
//func VerifyPasswordResetCode(db *sql.DB, email string, code string) (string, error) {
//	var userID string
//	var codeID string
//
//	err := db.QueryRow(`
//		SELECT id
//		FROM users
//		WHERE email = ?
//	`, email).Scan(&userID)
//
//	if err != nil {
//		return "", errors.New("code incorrect ou expiré")
//	}
//
//	codeHash := HashTwoFactorCode(code)
//
//	err = db.QueryRow(`
//		SELECT id
//		FROM password_reset_codes
//		WHERE user_id = ?
//		AND code_hash = ?
//		AND expires_at > datetime('now')
//		AND used = 0
//		ORDER BY expires_at DESC
//		LIMIT 1
//	`, userID, codeHash).Scan(&codeID)
//
//	if err != nil {
//		return "", errors.New("code incorrect ou expiré")
//	}
//
//	_, err = db.Exec(`
//		UPDATE password_reset_codes
//		SET used = 1
//		WHERE id = ?
//		AND used = 0
//	`, codeID)
//
//	if err != nil {
//		return "", err
//	}
//
//	return userID, nil
//}
//
//func ResetPassword(db *sql.DB, userID string, newPassword string) error {
//	if newPassword == "" {
//		return errors.New("mot de passe obligatoire")
//	}
//
//	hash, err := bcrypt.GenerateFromPassword(
//		[]byte(newPassword),
//		bcrypt.DefaultCost,
//	)
//
//	if err != nil {
//		return err
//	}
//
//	_, err = db.Exec(`
//		UPDATE users
//		SET password_hash = ?
//		WHERE id = ?
//	`, string(hash), userID)
//
//	return err
//}
//
//func SendPasswordResetEmail(destination string, code string) error {
//	smtpEmail := os.Getenv("SMTP_EMAIL")
//	smtpPassword := os.Getenv("SMTP_PASSWORD")
//
//	if smtpEmail == "" || smtpPassword == "" {
//		return fmt.Errorf("SMTP_EMAIL ou SMTP_PASSWORD manquant")
//	}
//
//	auth := smtp.PlainAuth(
//		"",
//		smtpEmail,
//		smtpPassword,
//		"smtp.gmail.com",
//	)
//
//	message := []byte(
//		"From: " + smtpEmail + "\r\n" +
//			"To: " + destination + "\r\n" +
//			"Subject: Réinitialisation du mot de passe - Forum\r\n" +
//			"MIME-Version: 1.0\r\n" +
//			"Content-Type: text/plain; charset=UTF-8\r\n" +
//			"\r\n" +
//			"Votre code de réinitialisation est : " + code + "\r\n\r\n" +
//			"Ce code est valable pendant 5 minutes.\r\n" +
//			"Si vous n'êtes pas à l'origine de cette demande, ignorez cet email.\r\n",
//	)
//
//	return smtp.SendMail(
//		"smtp.gmail.com:587",
//		auth,
//		smtpEmail,
//		[]string{destination},
//		message,
//	)
//}
//
//func CleanupPasswordResetCodes(db *sql.DB) {
//	_, _ = db.Exec(`
//		DELETE FROM password_reset_codes
//		WHERE expires_at <= ?
//	`, time.Now())
//}
