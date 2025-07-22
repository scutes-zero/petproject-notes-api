package storage

import (
	"fmt"

	"github.com/mattn/go-sqlite3"

	"notes-api/internal/models"

	"database/sql"
)

func (s *Storage) CreateUser(username, password string) (int64, error) {
	const op = "storage.CreateUser"

	stmt, err := s.db.Prepare(`
		INSERT INTO users (username, password)
		VALUES (?, ?);
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, password)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, ErrUserAlreadyExists
		}

		return 0, fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) UserExists(username string) (bool, error) {
	const op = "storage.UserExists"

	stmt, err := s.db.Prepare(`
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE username = ?);
	`)
	if err != nil {
		return false, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	return exists, nil
}

func (s *Storage) User(username string) (*models.User, error) {
	const op = "storage.User"

	stmt, err := s.db.Prepare(`
		SELECT id, username, password
		FROM users
		WHERE username = ?;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var user models.User
	err = stmt.QueryRow(username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: user not found", op)
		}
		return nil, fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	return &user, nil
}
