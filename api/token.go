package api

import (
	"bootdev/token"
	"bootdev/utils"
	"net/http"
	"strconv"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	type refreshResponse struct {
		Token string `json:"token,omitempty"`
	}

	refreshToken, err := token.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	rToken, err := token.VerifyToken(refreshToken, refreshIssuer)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	isTokenRevoked, err := db.IsRevoked(refreshToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	if isTokenRevoked {
		utils.RespondWithError(w, http.StatusUnauthorized, "token revoked")
		return
	}

	idStr, err := rToken.Claims.GetSubject()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	id, _ := strconv.Atoi(idStr)
	t, err := token.CreateToken(accessTokenExpiry, id, accessIssuer)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, refreshResponse{t})
}

func RevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := token.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = token.VerifyToken(refreshToken, refreshIssuer)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	err = db.RevokeToken(refreshToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, nil)
}
