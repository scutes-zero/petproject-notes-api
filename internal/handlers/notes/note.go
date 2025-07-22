package notes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"notes-api/internal/models"
	store "notes-api/internal/storage"
	"notes-api/internal/utils"
	"notes-api/pkg/logger"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type NoteProvider interface {
	Note(id, userID int) (*models.Note, error)
}

func NoteHandler(log *slog.Logger, storage NoteProvider) http.HandlerFunc {
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

		note, err := storage.Note(id, userIDInt)
		if err != nil {
			if errors.Is(err, store.ErrNoteNotFound) {
				log.Warn("note not found", logger.Err(err))

				w.WriteHeader(http.StatusNotFound)
				encoder.Encode(map[string]string{"NotFound": "Note not found"})
				return
			}

			log.Error("error when retrieving note", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"InternalError": "Failed to retrieve note"})
			return
		}

		encoder.Encode(note)
	}
}
