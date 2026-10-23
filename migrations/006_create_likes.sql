CREATE TABLE IF NOT EXISTS likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    post_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    reaction TEXT NOT NULL DEFAULT 'like',

    UNIQUE(user_id, post_id),
    CHECK (reaction IN ('like', 'dislike')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);
