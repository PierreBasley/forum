package main

import (
	"database/sql"
	"forum/server"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db, err := sql.Open("sqlite3", "file:test.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Table des utilisateurs
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		)
	`)

	if err != nil {
		log.Fatal(err)
	}

	// Table des posts
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			categories TEXT,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)

	if err != nil {
		log.Fatal(err)
	}

	// Table des codes 2FA
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS two_factor_codes (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			code_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)

	if err != nil {
		log.Fatal(err)
	}

	server.StartHTTPServer(db)
}
