package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"notes-api/internal/models"

	"golang.org/x/crypto/bcrypt"

	"log/slog"

	"notes-api/pkg/logger"
)

type UserProvider interface {
	User(username string) (*models.User, error)
}

func LoginHandler(log *slog.Logger, storage UserProvider, jwtSecret []byte) http.HandlerFunc {
	type response struct {
		Token string `json:"token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req models.User

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request body", logger.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{"InvalidRequest": "Failed to decode request body"})

			return
		}

		if errs := req.Validate(); len(errs) > 0 {
			log.Error("validation error", logger.Err(fmt.Errorf("invalid user data: %v", errs)))

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{"ValidationError": fmt.Sprintf("Invalid user data: %v", errs)})

			return
		}

		user, err := storage.User(req.Username)
		if err != nil {
			log.Error("failed to retrieve user", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"StorageError": "Failed to retrieve user"})
			return
		}

		if err := verifyPassword(user.Password, req.Password); err != nil {
			log.Error("authentication failed", logger.Err(err))

			w.WriteHeader(http.StatusUnauthorized)
			encoder.Encode(map[string]string{"AuthenticationError": "Invalid username or password"})
			return
		}

		token, err := GenerateJWT(jwtSecret, user.ID)
		if err != nil {
			log.Error("failed to generate JWT token", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"TokenError": "Failed to generate JWT token"})
			return
		}

		w.WriteHeader(http.StatusOK)
		encoder.Encode(response{Token: token})
	}

}

func verifyPassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("error verifying password: %w", err)
	}

	return nil
}
