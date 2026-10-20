package server

import (
	"database/sql"
	"log"
)

func CreationDB() (*sql.DB, error) {

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

	return db, nil
}
