package token

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret               = []byte(os.Getenv("JWT_SECRET"))
	ErrNoAuthHeaderIncluded = errors.New("not auth header included in request")
)

func CreateToken(expiry, id int) (string, error) {
	now := time.Now()
	expires := now.Add(24 * time.Hour)
	if expiry > 0 && now.Add(time.Duration(expiry)*time.Second).Before(expires) {
		expires = now.Add(time.Duration(expiry))
	}

	// create token
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
			Subject:   fmt.Sprintf("%d", id),
		},
	)

	return t.SignedString(jwtSecret)
}

func VerifyToken(acessToken string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(
		acessToken,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return jwtSecret, nil },
	)
	if err != nil {
		log.Print("ParseWithClaims: ", err)
		return &jwt.Token{}, err
	}

	if !token.Valid {
		return &jwt.Token{}, err
	}

	return token, nil
}

// GetBearerToken -
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}
