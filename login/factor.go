package login

//
//import (
//	"crypto/rand"
//	"crypto/sha256"
//	"database/sql"
//	"encoding/hex"
//	"fmt"
//	"math/big"
//	"net/smtp"
//	"os"
//)
//
//// GenerateTwoFactorCode génère un code numérique de 6 chiffres.
//func GenerateTwoFactorCode() (string, error) {
//	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
//	if err != nil {
//		return "", err
//	}
//
//	return fmt.Sprintf("%06d", n.Int64()), nil
//}
//
//// HashTwoFactorCode transforme le code 2FA en hash SHA-256.
//func HashTwoFactorCode(code string) string {
//	hash := sha256.Sum256([]byte(code))
//	return hex.EncodeToString(hash[:])
//}
//
//// SendTwoFactorEmail envoie le code 2FA par SMTP Gmail.
//func SendTwoFactorEmail(destination string, code string) error {
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
//			"Subject: Code de verification - Forum\r\n" +
//			"MIME-Version: 1.0\r\n" +
//			"Content-Type: text/plain; charset=UTF-8\r\n" +
//			"\r\n" +
//			"Votre code de verification est : " + code + "\r\n\r\n" +
//			"Ce code est valable pendant 5 minutes.\r\n" +
//			"Si vous n'etes pas a l'origine de cette inscription, ignorez cet email.\r\n",
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
//func IsTwoFactorBlocked(db *sql.DB, userID string) (bool, error) {
//	var blocked int
//
//	err := db.QueryRow(`
//		SELECT blocked
//		FROM two_factor_attempts
//		WHERE user_id = ?
//	`, userID).Scan(&blocked)
//
//	if err == sql.ErrNoRows {
//		return false, nil
//	}
//
//	if err != nil {
//		return false, err
//	}
//
//	return blocked == 1, nil
//}
//
//func AddTwoFactorAttempt(db *sql.DB, userID string) (int, error) {
//	_, err := db.Exec(`
//		INSERT INTO two_factor_attempts (user_id, attempts, blocked)
//		VALUES (?, 1, 0)
//		ON CONFLICT(user_id)
//		DO UPDATE SET attempts = attempts + 1
//	`, userID)
//
//	if err != nil {
//		return 0, err
//	}
//
//	var attempts int
//
//	err = db.QueryRow(`
//		SELECT attempts
//		FROM two_factor_attempts
//		WHERE user_id = ?
//	`, userID).Scan(&attempts)
//
//	if err != nil {
//		return 0, err
//	}
//
//	if attempts >= 5 {
//		_, err = db.Exec(`
//			UPDATE two_factor_attempts
//			SET blocked = 1
//			WHERE user_id = ?
//		`, userID)
//
//		if err != nil {
//			return 0, err
//		}
//	}
//
//	return attempts, nil
//}
//
//func ResetTwoFactorAttempts(db *sql.DB, userID string) error {
//	_, err := db.Exec(`
//		DELETE FROM two_factor_attempts
//		WHERE user_id = ?
//	`, userID)
//
//	return err
//}
