package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"example.com/chirpy/internal/auth"
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
	type response struct {
		chirp
	}

	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Print("Chirp not found")
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpId)
	if err != nil {
		log.Printf("Error fetching chirp from database: %s", err)
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	respondWithJson(w, http.StatusOK, response{
		chirp: chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		},
	})
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	var dbChirps []database.Chirp
	authorId := r.URL.Query().Get("author_id")
	if authorId != "" {
		userId, err := uuid.Parse(authorId)
		if err != nil {
            log.Printf("Unable to parse author ID into UUID: %s", err)
			respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		dbChirpsByUser, err := cfg.dbQueries.GetAllChirpsByUserId(r.Context(), userId)
		if err != nil {
			log.Printf("Unable to fetch chirps from database: %s", err)
			respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		dbChirps = append(dbChirps, dbChirpsByUser...)
	} else {
		allDbChirps, err := cfg.dbQueries.GetAllChirps(r.Context())
		if err != nil {
			log.Printf("Unable to fetch chirps from database: %s", err)
			respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		dbChirps = append(dbChirps, allDbChirps...)
	}

	chirps := []chirp{}
	for _, dbChirp := range dbChirps {
		chirp := chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}
		chirps = append(chirps, chirp)
	}

	respondWithJson(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type response struct {
		chirp
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Unable to fetch bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userId, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("Invalid token: %s", err)
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
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
		UserID: userId,
		Body:   cleanedBody,
	})

	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Unable to create chirp")
	}

	respondWithJson(w, http.StatusCreated, response{
		chirp: chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		},
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Unable to get bearer token: %s", err)
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userId, err := auth.ValidateJWT(bearer, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Unable to get chirpId from path value: %s", err)
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpId)
	if err != nil {
		log.Printf("Error fetching chirp from database: %s", err)
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	if dbChirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	if err := cfg.dbQueries.DeleteChirp(r.Context(), chirpId); err != nil {
		log.Printf("Unable to delete chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	respondWithJson(w, http.StatusNoContent, http.StatusText(http.StatusNoContent))
}
