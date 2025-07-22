package auth

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"notes-api/internal/models"
	"notes-api/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type UserCreator interface {
	CreateUser(username, password string) (int64, error)
	UserExists(username string) (bool, error)
}

func RegisterHandler(log *slog.Logger, storage UserCreator) http.HandlerFunc {
	type response struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Error("failed to decode request body", logger.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{"InvalidRequest": "Failed to decode request body"})

			return
		}

		if errs := user.Validate(); len(errs) > 0 {
			log.Error("Validation error", logger.Err(fmt.Errorf("invalid user data: %v", errs)))

			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(map[string]string{"ValidationError": fmt.Sprintf("Invalid user data: %v", errs)})

			return
		}

		exists, err := storage.UserExists(user.Username)
		if err != nil {
			log.Error("failed to check if user exists", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"StorageError": "Failed to check if user exists"})

			return
		}

		if exists {
			log.Warn("user already exists", slog.String("username", user.Username))

			w.WriteHeader(http.StatusConflict)
			encoder.Encode(map[string]string{"UserExists": "User already exists"})

			return
		}

		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			log.Error("failed to hash password", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"HashingError": "Failed to hash password"})

			return
		}

		id, err := storage.CreateUser(user.Username, hashedPassword)
		if err != nil {
			log.Error("failed to create user", logger.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(map[string]string{"StorageError": "Failed to create user"})

			return
		}

		w.WriteHeader(http.StatusCreated)
		encoder.Encode(response{ID: id, Username: user.Username})
	}
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	return string(hashedPassword), nil
}
