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

	// SQL (requests and initialization) should be written in script files,
	// then read from
	// Ideally, prepared requests should be made
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

	// Initially, this was executed on a different thread (using go statement)
	server.StartHTTPServer(db)
}
