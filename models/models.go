package models

import (
	"forum/session"
	"time"
)

type User struct {
	ID        string
	Username  string
	Email     string
	CreatedAt time.Time
	Role      string
}

type Post struct {
	ID         string
	UserID     string
	Username   string
	Title      string
	Content    string
	Categories []string
	CreatedAt  time.Time
}

type PageData struct {
	Auth    session.Authentified
	Posts   []Post
	Profile User
}
