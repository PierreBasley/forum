package models

import (
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
	ID            string
	UserID        string
	Username      string
	Title         string
	Content       string
	Categories    []string
	CreatedAt     time.Time
	LikesCount    int
	LikedByMe     bool
	DislikesCount int
	DislikedByMe  bool
}

type Comment struct {
	Username  string
	Content   string
	CreatedAt time.Time
}

type PostsData struct {
	Posts    []Post
	Comments []Comment
}
