package main

import (
	"log"
	"net/http"

	"example.com/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Unable to fetch bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if err := cfg.dbQueries.RevokeRefreshToken(r.Context(), bearer); err != nil {
		log.Printf("Unable to revoke refresh token: %s", err)
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpdateRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Unable to fetch bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userId, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), bearer)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	newAccessToken, err := auth.MakeJWT(userId, cfg.jwtSecret, accessTokenExpiresIn)
	if err != nil {
		log.Printf("Unable to create new access token: %s", err)
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	respondWithJson(w, http.StatusOK, response{newAccessToken})
}
