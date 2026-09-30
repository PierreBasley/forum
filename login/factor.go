package login

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
)

// GenerateTwoFactorCode génère un code numérique de 6 chiffres.
func GenerateTwoFactorCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashTwoFactorCode transforme le code 2FA en hash SHA-256.
func HashTwoFactorCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

// SendTwoFactorEmail envoie le code 2FA par SMTP Gmail.
func SendTwoFactorEmail(destination string, code string) error {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpEmail == "" || smtpPassword == "" {
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
			"Si vous n'etes pas a l'origine de cette inscription, ignorez cet email.\r\n",
	)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		smtpEmail,
		[]string{destination},
		message,
	)
}
