package notes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"notes-api/internal/models"
	"notes-api/internal/utils"
	"notes-api/pkg/logger"
	"strconv"

	"log/slog"
)

type NoteCreator interface {
	CreateNote(userID int, title, content string) (int64, error)
}

func CreateNoteHandler(log *slog.Logger, storage NoteCreator) http.HandlerFunc {
	type response struct {
		ID      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var note models.Note

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			log.Error("failed to decode request body", logger.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{"InvalidRequest": "Failed to decode request body"})

			return
		}

		if errs := note.Validate(); len(errs) > 0 {
			log.Error("validation error", logger.Err(fmt.Errorf("invalid note data: %v", errs)))

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{"ValidationError": fmt.Sprintf("Invalid note data: %v", errs)})

			return
		}

		userID, ok := r.Context().Value(utils.UserIDKey).(string)
		if !ok {
			log.Error("user ID not found in context", logger.Err(fmt.Errorf("user ID not found in context")))

			w.WriteHeader(http.StatusUnauthorized)
			encoder.Encode(map[string]string{"Unauthorized": "User ID not found in context"})

			return
		}

		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			log.Error("invalid user ID", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"InvalidUserID": "Invalid user ID"})

			return
		}
		id, err := storage.CreateNote(userIDInt, note.Title, note.Content)
		if err != nil {
			log.Error("error creating note", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"Error": "Error creating note"})

			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response{
			ID:      id,
			Title:   note.Title,
			Content: note.Content,
		})
	}
}
