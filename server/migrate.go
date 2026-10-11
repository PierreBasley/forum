package server

import (
	"database/sql"
	"os"
	"strings"
)

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL
	)
	`)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		name := entry.Name()

		var count int
		err = db.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE filename = ?",
			name,
		).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		content, err := os.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return err
		}

		_, err = db.Exec(
			"INSERT INTO schema_migrations (filename, applied_at) VALUES (?, datetime('now'))",
			name,
		)
		if err != nil {
			return err
		}

	}

	return nil
}
