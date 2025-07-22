package notes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"strconv"

	"log/slog"

	"encoding/json"
	"fmt"
	"notes-api/internal/utils"
	"notes-api/pkg/logger"
)

type NoteDeleter interface {
	DeleteNote(id, userID int) error
}

func DeleteNoteHandler(log *slog.Logger, storage NoteDeleter) http.HandlerFunc {
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

		err = storage.DeleteNote(id, userIDInt)
		if err != nil {
			log.Error("error when deleting note", logger.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"InternalError": "Failed to delete note"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
