package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db, err := sql.Open("sqlite3", "file:test.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

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

	go startHTTPServer(db)

	fmt.Println("          FORUM")

	for {

		fmt.Println()
		fmt.Println("1. Se connecter")
		fmt.Println("2. Créer un compte")
		fmt.Println("3. Quitter")
		fmt.Println()

		var choice string

		fmt.Print("Choix : ")
		fmt.Scanln(&choice)

		switch choice {

		case "1":
			login(db)

		case "2":
			register(db)

		case "3":
			fmt.Println("Au revoir.")
			return

		default:
			fmt.Println("Choix invalide.")
		}
	}
}

func register(db *sql.DB) {

	fmt.Println()
	fmt.Println("        CRÉER UN COMPTE")

	var username string
	var email string
	var password string

	fmt.Print("Nom d'utilisateur : ")
	fmt.Scanln(&username)

	fmt.Print("Email : ")
	fmt.Scanln(&email)

	fmt.Print("Mot de passe : ")
	fmt.Scanln(&password)

	if strings.TrimSpace(username) == "" ||
		strings.TrimSpace(email) == "" ||
		strings.TrimSpace(password) == "" {

		fmt.Println("Tous les champs sont obligatoires.")
		return
	}

	err := CreateAccount(db, username, email, password)

	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}

	fmt.Println("Compte créé avec succès.")
}

func login(db *sql.DB) {

	fmt.Println()
	fmt.Println("          CONNEXION")

	var email string
	var password string

	fmt.Print("Email : ")
	fmt.Scanln(&email)

	fmt.Print("Mot de passe : ")
	fmt.Scanln(&password)

	if strings.TrimSpace(email) == "" ||
		strings.TrimSpace(password) == "" {

		fmt.Println("Email et mot de passe obligatoires.")
		return
	}

	_, err := Login(db, email, password)

	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}

	fmt.Println("Vous êtes connecté.")
}
