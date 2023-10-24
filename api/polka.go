package api

import (
	"bootdev/database"
	"bootdev/token"
	"bootdev/utils"
	"encoding/json"
	"net/http"
)

func UpgradeUser(w http.ResponseWriter, r *http.Request) {
	type upgradeRequest struct {
		Event string `json:"event,omitempty"`
		Data  struct {
			UserId int `json:"user_id,omitempty"`
		} `json:"data,omitempty"`
	}

	isApiKeyValid, err := token.VerifyApiKey(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if !isApiKeyValid {
		utils.RespondWithError(w, http.StatusUnauthorized, "api key invalid")
		return
	}

	decoder := json.NewDecoder(r.Body)

	uReq := upgradeRequest{}
	err = decoder.Decode(&uReq)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "cannot decode request body")
		return
	}

	if uReq.Event != "user.upgraded" {
		utils.RespondWithJSON(w, http.StatusOK, nil)
		return
	}

	u, err := db.UpdateUser(uReq.Data.UserId, "", "", true)
	if err == database.ErrNotFound {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, u)
}
