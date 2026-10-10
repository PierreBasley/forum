package server

import (
	"database/sql"
	"log"
)

func CreationDB() (*sql.DB, error) {
	// YF: SQL scripts should be read directly from a file and executed, not hardcoded

	//Creation or opening of the database "db.sqlite"
	db, err := sql.Open("sqlite3", "bdd/db.sqlite")
	if err != nil {
		return db, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	//Use of PRAGMA command to enable the foreign keys which are disabled by default
	_, err = db.Exec(`
	PRAGMA foreign_keys = ON;
	`)
	if err != nil {
		log.Println("Problème dans la commande PRAGMA")
		return db, err
	}

	// Table des codes 2FA
	// _, err = db.Exec(`
	// 	CREATE TABLE IF NOT EXISTS two_factor_codes (
	// 		id TEXT PRIMARY KEY,
	// 		user_id TEXT NOT NULL,
	// 		code_hash TEXT NOT NULL,
	// 		expires_at DATETIME NOT NULL,
	// 		used INTEGER NOT NULL DEFAULT 0,
	// 		FOREIGN KEY (user_id) REFERENCES users(id)
	// 	)
	// `)

	// Table des tentatives 2FA
	/*_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS two_factor_attempts (
			user_id TEXT PRIMARY KEY,
			attempts INTEGER NOT NULL DEFAULT 0,
			blocked INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)*/

	/*if err != nil {
			log.Fatal(err)
		}

		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS password_reset_codes (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			code_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)*/

	//Need to be removed, not the best method to insert an admin account.
	// var exists bool
	// err = db.QueryRow(`
	// 	SELECT EXISTS(SELECT 1 FROM users WHERE username = 'Root' AND email = 'root@caramail.com')
	// ̀`).Scan(&exists)
	// if err != nil {
	// 	return db, err
	// }
	// if exists == true {
	// 	log.Println("Compte Root déjà créé")
	// } else {
	// 	err = rootAccount(db)
	// 	if err != nil {
	// 		log.Println("Erreur dans la création du compte Root")
	// 	}
	// }

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY, 
		user_id TEXT NOT NULL, 
		created_at TEXT NOT NULL,
		expires_at TEXT NOT NULL
	)
	`)
	if err != nil {
		log.Println("Erreur création sessions :", err)
		return db, err
	}

	return db, nil
}

//Need to be removed, not the best method to insert an admin account.
// func rootAccount(db *sql.DB) error {

// 	username := "Root"
// 	email := "root@caramail.com"
// 	password := "rootTest12345"

// 	err := login.CreateAccount(db, username, email, password)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
