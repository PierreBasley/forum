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

	// _, err = db.Exec(`
	// CREATE TABLE IF NOT EXISTS sessions (
	// 	id TEXT PRIMARY KEY,
	// 	user_id TEXT NOT NULL,
	// 	created_at TEXT NOT NULL,
	// 	expires_at TEXT NOT NULL
	// )
	// `)
	// if err != nil {
	// 	log.Println("Erreur création sessions :", err)
	// 	return db, err
	// }

	return db, nil
}
