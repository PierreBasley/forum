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

	//Creation of the differents tables thanks to the following commands

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE
	)
	`)
	if err != nil {
		return db, err
	}
	//implementation of the different roles inside the DB
	_, err = db.Exec(`
	INSERT INTO roles(name) VALUES
	('visitor'),
	('user'),
	('moderator'),
	('admin')
	ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		log.Println("impossible de remplir la table role")
		return db, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username VARCHAR(100) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		created_at TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		roles_id INTEGER NOT NULL,
		FOREIGN KEY (roles_id) REFERENCES roles(id)
	)
	`)
	if err != nil {
		log.Println("Problème dans la création de la table users")
		return db, err
	}

	// Table des posts
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			categories TEXT,
			created_at TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		log.Println("Problème dans la création de la table posts")
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

	/*
		if err != nil {
			log.Fatal(err)
		}

		if err != nil {
			log.Fatal(err)
		}
	*/
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS password_reset_codes (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			code_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)

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
