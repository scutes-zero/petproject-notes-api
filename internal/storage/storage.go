package storage

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrNoteNotFound = errors.New("note not found")

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open database: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: failed to enable foreign keys: %w", op, err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        );`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: failed to create users table: %w", op, err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS notes (
            id INTEGER PRIMARY KEY,
            user_id INTEGER NOT NULL,
            title TEXT NOT NULL,
            content TEXT,
            created_at TEXT NOT NULL DEFAULT current_timestamp,
            updated_at TEXT NOT NULL DEFAULT current_timestamp,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: failed to create notes table: %w", op, err)
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: failed to create index: %w", op, err)
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_notes_id_user_id ON notes(id, user_id);")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: failed to create index: %w", op, err)
	}

	return &Storage{db: db}, nil
}
