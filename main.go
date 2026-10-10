package main

import (
	"forum/server"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

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
	// Initially, this was executed on a different thread (using go statement)
	server.StartHTTPServer(db)
}
