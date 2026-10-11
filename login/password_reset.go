package login

import (
	"database/sql"
	"errors"

	"net/smtp"
	"os"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreatePasswordResetCode(db *sql.DB, email string) error {
	var userID string

	err := db.QueryRow(
		"SELECT id FROM users WHERE email = ?",
		email,
	).Scan(&userID)
	if err != nil {
		return errors.New("email introuvable")
	}

	code, err := GenerateTwoFactorCode()
	if err != nil {
		return err
	}

	codeID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	codeHash := HashTwoFactorCode(code)

	_, err = db.Exec(
		"DELETE FROM password_reset_codes WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO password_reset_codes
		(id, user_id, code_hash, expires_at, used)
		VALUES (?, ?, ?, datetime('now', '+5 minutes'), 0)
	`, codeID.String(), userID, codeHash)
	if err != nil {
		return err
	}

	if err := SendPasswordResetEmail(email, code); err != nil {
		_, _ = db.Exec(
			"DELETE FROM password_reset_codes WHERE id = ?",
			codeID.String(),
		)
		return err
	}

	return nil
}

func VerifyPasswordResetCode(db *sql.DB, code string) (string, error) {
	var userID string

	//TODO : There is a problem about that request. Where is the email variable ?
	err := db.QueryRow(
		"SELECT id FROM users WHERE email = ?",
	).Scan(&userID)

	if err != nil {
		return "", errors.New("code incorrect ou expiré")
	}

	var codeID string

	err = db.QueryRow(`
		SELECT id
		FROM password_reset_codes
		WHERE user_id = ?
		AND code_hash = ?
		AND expires_at > datetime('now')
		AND used = 0
		LIMIT 1
	`, userID, HashTwoFactorCode(code)).Scan(&codeID)

	if err != nil {
		return "", errors.New("code incorrect ou expiré")
	}

	_, err = db.Exec(
		"UPDATE password_reset_codes SET used = 1 WHERE id = ?",
		codeID,
	)

	if err != nil {
		return "", err
	}

	return userID, nil
}

func ResetPassword(db *sql.DB, userID, password string) error {
	if !IsValidPassword(password) {
		return errors.New(
			"Le mot de passe doit contenir au minimum 8 caractères, 1 chiffre, 1 majuscule et 1 caractère spécial",
		)
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE users SET password_hash = ? WHERE id = ?",
		string(hash),
		userID,
	)

	return err
}

func SendPasswordResetEmail(email, code string) error {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpEmail == "" || smtpPassword == "" {
		return errors.New("SMTP_EMAIL ou SMTP_PASSWORD manquant")
	}

	auth := smtp.PlainAuth(
		"",
		smtpEmail,
		smtpPassword,
		"smtp.gmail.com",
	)

	message := []byte(
		"From: " + smtpEmail + "\r\n" +
			"To: " + email + "\r\n" +
			"Subject: Réinitialisation du mot de passe - Forum\r\n" +
			"\r\n" +
			"Votre code est : " + code + "\r\n" +
			"Il est valable pendant 5 minutes.\r\n",
	)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		smtpEmail,
		[]string{email},
		message,
	)
}
