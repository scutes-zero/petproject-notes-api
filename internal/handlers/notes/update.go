package notes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"strconv"

	"log/slog"

	"encoding/json"

	"notes-api/internal/models"
	"notes-api/internal/utils"

	"fmt"

	"notes-api/pkg/logger"
)

type NoteUpdater interface {
	UpdateNote(id, userID int, title, content string) error
}

func UpdateNoteHandler(log *slog.Logger, storage NoteUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("error when converting id to int", logger.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{"InvalidID": "ID must be an integer"})
			return
		}

		var note models.Note
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			log.Error("failed to decode request body", logger.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{"InvalidRequest": "Failed to decode request body"})
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
			log.Error("error when converting user ID to int", logger.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"InternalError": "Failed to convert user ID"})
			return
		}

		if err := storage.UpdateNote(id, userIDInt, note.Title, note.Content); err != nil {
			log.Error("error when updating note", logger.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"InternalError": "Failed to update note"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
