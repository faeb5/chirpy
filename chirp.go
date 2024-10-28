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

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
    chirpId, err := uuid.Parse(r.PathValue("chirpID"))
    if err != nil {
        log.Print("Chirp not found")
        respondWithError(w, http.StatusNotFound, "Chirp not found")
        return
    }

    dbChirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpId)
    if err != nil {
        log.Printf("Error fetching chirp from database: %s", err)
        respondWithError(w, http.StatusInternalServerError, "Something went wrong")
        return
    }

    respondWithJson(w, http.StatusOK, convertDatabaseChirp(dbChirp))
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
    dbChirps, err := cfg.dbQueries.GetAllChirps(r.Context())
    if err != nil {
        log.Printf("Error fetching all chirps from database: %s", err)
        respondWithError(w, http.StatusInternalServerError, "Something went wrong")
        return
    }

    var chirps []chirp
    for _, dbChirp := range dbChirps {
        chirp := convertDatabaseChirp(dbChirp)
        chirps = append(chirps, chirp)
    }

	respondWithJson(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
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
	dbChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: params.UserId,
		Body:   cleanedBody,
	})

	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Unable to create chirp")
	}

    chirp := convertDatabaseChirp(dbChirp)
	respondWithJson(w, http.StatusCreated, chirp)
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

func convertDatabaseChirp(dbChirp database.Chirp) chirp {
    return chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}
