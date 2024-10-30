package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/chirpy/internal/auth"
	"example.com/chirpy/internal/database"
	"github.com/google/uuid"
)

type user struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		user
	}

	decoder := json.NewDecoder(r.Body)
	var params parameters
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error parsing JSON parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing the password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	newUser, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		HashedPassword: hashedPassword,
		Email:          params.Email,
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Unable to create user")
		return
	}

	respondWithJson(w, http.StatusCreated, response{
		user: user{
			ID:        newUser.ID,
			CreatedAt: newUser.CreatedAt,
			UpdatedAt: newUser.UpdatedAt,
			Email:     newUser.Email,
		},
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type response struct {
		user
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	var params parameters
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error parsing JSON parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	dbUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error querying user by email: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	expiresIn := time.Hour
	if params.ExpiresInSeconds > 0 && params.ExpiresInSeconds < 3600 {
		expiresIn = time.Duration(params.ExpiresInSeconds) * time.Second
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("Unable to create JWT: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJson(w, http.StatusOK, response{
		user: user{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		},
		Token: token,
	})
}
