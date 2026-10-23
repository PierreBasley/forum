CREATE TABLE IF NOT EXISTS comment_likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    comment_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    reaction TEXT NOT NULL DEFAULT 'like',

    UNIQUE(user_id, comment_id),
    CHECK (reaction IN ('like', 'dislike')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
);
