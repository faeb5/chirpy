package main

import (
	"encoding/json"
	"log"
	"net/http"

	"example.com/chirpy/internal/auth"
	"github.com/google/uuid"
)

type userEvent string

const (
	userUpgraded userEvent = "user.upgraded"
)

func (cfg *apiConfig) handlerEnableChirpyRed(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event userEvent `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		}
	}

    apiKey, err := auth.GetAPIKey(r.Header)
    if err != nil {
        log.Printf("Unable to find API key in request: %s", err)
        respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
        return
    }

    if apiKey != cfg.polkaKey {
        log.Printf("Invalid API key: %s", err)
        respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
        return
    }

	decoder := json.NewDecoder(r.Body)
	var params parameters
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Unable to decode JSON paramters: %s", err)
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if params.Event != userUpgraded {
		respondWithJson(w, http.StatusNoContent, http.StatusText(http.StatusNoContent))
		return
	}

	userId, err := uuid.Parse(params.Data.UserId)
	if err != nil {
		log.Printf("Unable to parse user ID: %s", err)
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if err := cfg.dbQueries.EnableChirpyRed(r.Context(), userId); err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
