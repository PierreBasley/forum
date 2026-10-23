CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

INSERT INTO roles(name) VALUES
('user'),
('moderator'),
('admin')
ON CONFLICT DO NOTHING;