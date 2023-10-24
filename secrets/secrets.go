package secrets

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	jwtSecret []byte
	apiKey    string
)

type secrets struct {
	JwtSecret []byte
	ApiKey    string
}

var keys secrets

func GetSecret() secrets {
	if keys.JwtSecret == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}

		keys.JwtSecret = []byte(os.Getenv("JWT_SECRET"))
		keys.ApiKey = os.Getenv("API_KEY")
	}

	return keys
}
