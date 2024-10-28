package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	type returnValues struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding chirp parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error decoding JSON")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedBody := replaceProfanities(params.Body)
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: params.UserId,
		Body:   cleanedBody,
	})

	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Unable to create chirp")
	}

	respondWithJson(w, http.StatusCreated, returnValues{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func replaceProfanities(text string) string {
	profanityMap := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}
	textFields := strings.Fields(text)
	for i, word := range textFields {
		lowerWord := strings.ToLower(word)
		if v, ok := profanityMap[lowerWord]; ok {
			textFields[i] = v
		}
	}
	return strings.Join(textFields, " ")
}
