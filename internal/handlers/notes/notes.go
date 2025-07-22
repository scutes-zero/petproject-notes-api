package notes

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"notes-api/internal/models"
	"notes-api/internal/utils"
	"strconv"

	"notes-api/pkg/logger"

	"errors"
	store "notes-api/internal/storage"
)

type NotesProvider interface {
	Notes(userID int) ([]models.Note, error)
}

func NotesHandler(log *slog.Logger, storage NotesProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

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

		notes, err := storage.Notes(userIDInt)
		if err != nil {
			if errors.Is(err, store.ErrNoteNotFound) {
				log.Warn("no notes found for user", logger.Err(err))
				w.WriteHeader(http.StatusNotFound)
				encoder.Encode(map[string]string{"NotFound": "No notes found for user"})
				return
			}

			log.Error("error when retrieving notes", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"InternalError": "Failed to retrieve notes"})
			return
		}

		encoder.Encode(notes)
	}
}
