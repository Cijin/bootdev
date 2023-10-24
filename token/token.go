package token

import (
	"bootdev/secrets"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	keys                    = secrets.GetSecret()
	jwtSecret               = keys.JwtSecret
	apiKey                  = keys.ApiKey
	ErrNoAuthHeaderIncluded = errors.New("not auth header included in request")
)

func CreateToken(expiry time.Duration, id int, issuer string) (string, error) {
	now := time.Now()
	expires := now.Add(expiry)

	// create token
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
			Subject:   fmt.Sprintf("%d", id),
		},
	)

	return t.SignedString(jwtSecret)
}

func VerifyToken(token, issuerType string) (*jwt.Token, error) {
	t, err := jwt.ParseWithClaims(
		token,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return jwtSecret, nil },
	)
	if err != nil {
		log.Print("ParseWithClaims: ", err)
		return &jwt.Token{}, err
	}

	if !t.Valid {
		return &jwt.Token{}, errors.New("token is not valid")
	}

	issuer, err := t.Claims.GetIssuer()
	if err != nil {
		return &jwt.Token{}, err
	}

	if issuer != issuerType {
		return &jwt.Token{}, errors.New("invalid token type")
	}

	return t, nil
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

// VerifyApiKey -
func VerifyApiKey(headers http.Header) (bool, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return false, ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return false, errors.New("malformed authorization header")
	}

	return splitAuth[1] == apiKey, nil
}
