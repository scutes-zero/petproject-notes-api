package auth

import (
	"strconv"

	"github.com/golang-jwt/jwt/v5"

	"time"
)

func GenerateJWT(jwtSecret []byte, userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtSecret)
}
