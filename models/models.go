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
	Score         int
	RepliesCount  int
}

type Comment struct {
	ID            string
	Username      string
	Content       string
	CreatedAt     time.Time
	LikesCount    int
	LikedByMe     bool
	DislikesCount int
	DislikedByMe  bool
	Score         int
}

type PostsPageData struct {
	Auth               session.Authentified
	Posts              []Post
	Sort               string
	Search             string
	Stacks             []string
	SelectedCategories map[string]bool
	OnlyMine           bool
	OnlyLiked          bool
}

type PostPageData struct {
	Auth     session.Authentified
	Post     Post
	Comments []Comment
}

type PageData struct {
	Auth    session.Authentified
	Profile User
	Stacks  []string
}
