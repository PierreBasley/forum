package main

import (
	"forum/server"
	"forum/session"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const REFRESH_INTERVAL = 1 * time.Minute
const SESSIONS_DIR = "tmp/sessions"

func main() {

	//Initialize the db and the tables
	db, err := server.CreationDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = server.RunMigrations(db)
	if err != nil {
		log.Fatal(err)
	}

	//creation of a manager which will handle the sessions of the website
	m, err := session.NewManager(SESSIONS_DIR)
	if err != nil {
		log.Fatal("Erreur : ", err)
	}

	//Goroutine configuration to clean expired sessions every minutes
	go func() {
		ticker := time.NewTicker(REFRESH_INTERVAL)
		defer ticker.Stop()
		for range ticker.C {
			err := m.NettoyageSessionsExpirees()
			if err != nil {
				log.Println("Erreur lors du nettoyage des sessions :", err)
			}
		}
	}()

	// Initially, this was executed on a different thread (using go statement)
	server.StartHTTPServer(db, m)
}
