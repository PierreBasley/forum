package main

import (
	"forum/server"
	"forum/session"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const REFRESH_INTERVAL = 1 * time.Minute

func main() {

	//Initialize the db and the tables
	db, err := server.CreationDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	m := session.NewManager()

	//Configuration d'une goroutine
	go func() {
		ticker := time.NewTicker(REFRESH_INTERVAL)
		defer ticker.Stop()
		for range ticker.C {
			m.NettoyageSessionsExpirees()
		}
	}()
	// Initially, this was executed on a different thread (using go statement)
	server.StartHTTPServer(db, m)
}
