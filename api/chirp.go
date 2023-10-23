package api

import (
	"bootdev/database"
	"bootdev/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func CreateChirp(w http.ResponseWriter, r *http.Request) {
	bannedWords := []string{"kerfuffle", "sharbert", "fornax"}
	censor := "****"

	decoder := json.NewDecoder(r.Body)

	c := &database.Chirp{}

	err := decoder.Decode(&c)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(c.Body) > 140 {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	words := strings.Split(c.Body, " ")
	for i, w := range words {
		for _, bw := range bannedWords {
			if strings.EqualFold(w, bw) {
				words[i] = censor
			}
		}
	}
	censoredText := strings.Join(words, " ")

	chirp, err := db.CreateChirp(censoredText)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, chirp)
	return
}

func GetChirp(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, "Not a valid id")
		return
	}

	c, err := db.GetChirp(id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "Chirp does not exist")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, c)
}

func GetChrips(w http.ResponseWriter, r *http.Request) {
	chirps, err := db.GetChirps()
	if err != nil {
		log.Print(err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, chirps)
}
