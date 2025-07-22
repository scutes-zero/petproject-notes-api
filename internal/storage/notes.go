package storage

import (
	"fmt"
	"notes-api/internal/models"
)

func (s *Storage) CreateNote(userID int, title, content string) (int64, error) {
	const op = "storage.CreateNote"

	stmt, err := s.db.Prepare(`
		INSERT INTO notes (user_id, title, content)
		VALUES (?, ?, ?);
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID, title, content)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) Note(id, userID int) (*models.Note, error) {
	const op = "storage.Note"

	stmt, err := s.db.Prepare(`
		SELECT *
		FROM notes
		WHERE id = ? AND user_id = ?;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	row, err := stmt.Query(id, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}
	defer row.Close()

	var note models.Note
	if !row.Next() {
		if err := row.Err(); err != nil {
			return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
		}

		return nil, ErrNoteNotFound
	}

	if err := row.Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
		return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
	}

	return &note, nil
}

func (s *Storage) Notes(userID int) ([]models.Note, error) {
	const op = "storage.Notes"

	stmt, err := s.db.Prepare(`
		SELECT *
		FROM notes
		WHERE user_id = ?;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(notes) == 0 {
		return nil, ErrNoteNotFound
	}

	return notes, nil
}

func (s *Storage) DeleteNote(id, userID int) error {
	const op = "storage.DeleteNote"

	stmt, err := s.db.Prepare(`
		DELETE FROM notes
		WHERE id = ? AND user_id = ?;
	`)
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id, userID)
	if err != nil {
		return fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no rows deleted", op)
	}

	return nil
}

func (s *Storage) UpdateNote(id, userID int, title, content string) error {
	const op = "storage.UpdateNote"

	stmt, err := s.db.Prepare(`
		UPDATE notes
		SET title = ?, content = ?, updated_at = current_timestamp
		WHERE id = ? AND user_id = ?;
	`)
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(title, content, id, userID)
	if err != nil {
		return fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no rows updated", op)
	}

	return nil
}
