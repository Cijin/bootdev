package api

import (
	"bootdev/database"
	"bootdev/token"
	"bootdev/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

var db = database.GetDb()

func CreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	u := &database.User{}

	err := decoder.Decode(&u)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := db.CreateUser(u.Email, u.Password)
	if err != nil {
		if errors.Is(err, database.ErrDuplicateEmail) {
			utils.RespondWithError(w, http.StatusUnauthorized, "User with email already exists")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, user)
	return
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	accessToken, err := token.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	token, err := token.VerifyToken(accessToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	idStr, err := token.Claims.GetSubject()
	if err != nil {
		log.Print("Claims.GetSubject: ", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Print(err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	decoder := json.NewDecoder(r.Body)

	u := &database.User{}

	err = decoder.Decode(&u)
	if err != nil {
		log.Print(err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	res, err := db.UpdateUser(id, u.Email, u.Password)
	if err != nil {
		log.Print(err)
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, res)
	return
}

func Login(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email            string `json:"email,omitempty"`
		Password         string `json:"password,omitempty"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	}

	type loginResponse struct {
		database.User
		Token string `json:"token,omitempty"`
	}

	decoder := json.NewDecoder(r.Body)

	req := &loginRequest{}

	err := decoder.Decode(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := db.Login(req.Email, req.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ss, err := token.CreateToken(req.ExpiresInSeconds, user.Id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	res := loginResponse{user, ss}

	utils.RespondWithJSON(w, http.StatusOK, res)
	return
}
