package api

import (
	"bootdev/database"
	"bootdev/token"
	"bootdev/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

const censor = "****"

var bannedWords = []string{"kerfuffle", "sharbert", "fornax"}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func CreateChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := token.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	t, err := token.VerifyToken(accessToken, accessIssuer)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	idStr, err := t.Claims.GetSubject()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	id, _ := strconv.Atoi(idStr)

	decoder := json.NewDecoder(r.Body)

	c := &database.Chirp{}

	err = decoder.Decode(&c)
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

	chirp, err := db.CreateChirp(id, censoredText)
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

	aId := -1

	aIdStr := r.URL.Query().Get("author_id")
	sortBy := r.URL.Query().Get("sort")

	if aIdStr != "" {
		aId, err = strconv.Atoi(aIdStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "author_id is not valid id")
			return
		}
	}

	if aId != -1 {
		chirpsCopy := chirps
		chirps = []database.Chirp{}
		for _, c := range chirpsCopy {
			if aId == c.AuthorId {
				chirps = append(chirps, c)
			}
		}
	}

	sort.Slice(chirps, func(i, j int) bool {
		if sortBy != "" && sortBy == "desc" {
			return chirps[i].Id > chirps[j].Id
		}
		return chirps[i].Id < chirps[j].Id
	})

	utils.RespondWithJSON(w, http.StatusOK, chirps)
}

func DeleteChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := token.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	t, err := token.VerifyToken(accessToken, accessIssuer)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	uidStr, err := t.Claims.GetSubject()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	uId, _ := strconv.Atoi(uidStr)

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Not a valid id")
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

	if c.AuthorId != uId {
		utils.RespondWithError(w, http.StatusForbidden, "You are not allowed to do this")
		return
	}

	chirp, err := db.DeleteChirp(c.Id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, chirp)
}
