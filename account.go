package main

import (
	"database/sql"
	"fmt"

	"github.com/gofrs/uuid"
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
		return fmt.Errorf("cet email est déjà utilisé")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash)
		VALUES (?, ?, ?, ?)
	`, id.String(), username, email, hash)

	if err != nil {
		return err
	}

	fmt.Println("Compte créé avec succès")

	return nil
}

func Login(db *sql.DB, email string, password string) error {

	var hash string

	err := db.QueryRow(`
		SELECT password_hash
		FROM users
		WHERE email = ?
	`, email).Scan(&hash)

	if err != nil {
		return fmt.Errorf("email ou mot de passe incorrect")
	}

	if !CheckPasswordHash(password, hash) {
		return fmt.Errorf("email ou mot de passe incorrect")
	}

	fmt.Println("Connexion réussie")

	return nil
}
