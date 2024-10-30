package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const tokenIssuer string = "chirpy"

func HashPassword(password string) (string, error) {
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed_password), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userId.String(),
	})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)

	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Printf("Unable to get issuer from claims: %s", err)
		return uuid.Nil, err
	}
	if issuer != tokenIssuer {
		return uuid.Nil, errors.New("Invalid issuer")
	}

	expirationTime, err := token.Claims.GetExpirationTime()
	if err != nil {
		log.Printf("Unable to get expiration time from claims: %s", err)
		return uuid.Nil, err
	}
	if expirationTime.UTC().Before(time.Now().UTC()) {
		return uuid.Nil, errors.New("Token is expired")
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("Unable to retrieve subject from claims: %s", err)
	}

	return uuid.Parse(subject)
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader, ok := headers["Authorization"]
	if !ok {
		return "", fmt.Errorf("Authorization header not found")
	}

	var bearer string
	for _, val := range authHeader {
		fields := strings.Fields(val)
		if len(fields) != 2 {
			continue
		}
		if strings.ToLower(fields[0]) == "bearer" {
			bearer = fields[1]
		}
	}
	if bearer == "" {
		return "", fmt.Errorf("Bearer not found")
	}

	return bearer, nil
}

func MakeRefreshToken() (string, error) {
	rawData := make([]byte, 32)
	if _, err := rand.Read(rawData); err != nil {
		log.Printf("Unable to create raw data for refresh token")
		return "", errors.New("Unable to create refresh token")
	}
	return hex.EncodeToString(rawData), nil
}
