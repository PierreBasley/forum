package login

import (
	"database/sql"
	"errors"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateAccount(db *sql.DB, username string, email string, password string) error {

	id, err := uuid.NewV4()
	if err != nil {
		return err
	}

	var exists bool

	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)",
		email,
	).Scan(&exists)

	if err != nil {
		return err
	}
	if exists {
		return errors.New("email déjà utilisé")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash)
		VALUES (?, ?, ?, ?)
	`, id.String(), username, email, string(hash))

	if err != nil {
		return err
	}

	return nil
}

func Login(db *sql.DB, email string, password string) (string, error) {

	var userID string
	var hash string

	err := db.QueryRow(`
		SELECT id, password_hash
		FROM users
		WHERE email = ?
	`, email).Scan(&userID, &hash)

	if err != nil {
		return "", errors.New("email ou mot de passe incorrect")
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		return "", errors.New("email ou mot de passe incorrect")
	}

	return userID, nil
}
