CREATE TABLE IF NOT EXISTS posts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			categories TEXT,
			created_at TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);