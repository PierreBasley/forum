CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    roles_id INTEGER NOT NULL,
    FOREIGN KEY (roles_id) REFERENCES roles(id)
);