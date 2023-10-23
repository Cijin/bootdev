package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	type errRes struct {
		Error string `json:"error,omitempty"`
	}

	r := errRes{}

	r.Error = msg
	d, err := json.Marshal(r)
	if err != nil {
		log.Printf("Marshal error: %s", err.Error())
	}

	w.WriteHeader(code)
	w.Write(d)
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	d, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Marshal error: %s", err.Error())
	}

	w.WriteHeader(code)
	w.Write(d)
}
