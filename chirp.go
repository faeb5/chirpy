package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding JSON")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedBody := replaceProfanities(params.Body)
	respondWithJson(w, http.StatusOK, returnVals{cleanedBody})
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
