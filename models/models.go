package models

type User struct {
	ID         string
	Username   string
	Email      string
	Created_at string
	Role       string
}

type Post struct {
	ID         string
	UserID     string
	Username   string
	Title      string
	Content    string
	Categories []string
	CreatedAt  string
}
