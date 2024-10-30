package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	password1 := "superSecret123!"
	password2 := "anotherSecret321?"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: password1,
			hash:     "wrongPassword",
			wantErr:  true,
		},
		{
			name:     "Password with different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Incorrect hash",
			password: password1,
			hash:     "invalidHash",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CheckPasswordHash(test.password, test.hash)
			if (err != nil) != test.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userId := uuid.New()
	secret := "supersecret"
	tokenString, _ := MakeJWT(userId, secret, time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserId  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: tokenString,
			tokenSecret: secret,
			wantUserId:  userId,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "wrong.token.string",
			tokenSecret: secret,
			wantUserId:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Invalid secret",
			tokenString: tokenString,
			tokenSecret: "wrongSecret",
			wantUserId:  uuid.Nil,
			wantErr:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			uuid, err := ValidateJWT(test.tokenString, test.tokenSecret)
			if uuid != test.wantUserId {
				t.Errorf("ValidateJWT() uuid.UUID = %v, want %v", uuid, test.wantUserId)
			}
			if (err != nil) != test.wantErr {
				t.Errorf("ValidateJWT() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tokenString, _ := MakeJWT(uuid.New(), "supersecret", time.Hour)
	bearer := fmt.Sprintf("Bearer %s", tokenString)
	headers := http.Header{
		"Authorization": []string{bearer},
	}

	tests := []struct {
		name            string
		headers         http.Header
		wantBearerToken string
		wantErr         bool
	}{
		{
			name:            "Bearer Token found",
			headers:         headers,
			wantBearerToken: tokenString,
			wantErr:         false,
		},
		{
			name:            "Authorization header not found",
			headers:         http.Header{},
			wantBearerToken: "",
			wantErr:         true,
		},
		{
			name:            "Bearer not found",
			headers:         http.Header{"Authorization": []string{}},
			wantBearerToken: "",
			wantErr:         true,
		},
		{
			name:            "Invalid Bearer",
			headers:         http.Header{"Authorization": []string{"TotallyWrongBearer token"}},
			wantBearerToken: "",
			wantErr:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenString, err := GetBearerToken(test.headers)
			if tokenString != test.wantBearerToken {
				t.Errorf("GetBearerToken() string = %v, want %v", tokenString, test.wantBearerToken)
			}
			if (err != nil) != test.wantErr {
				t.Errorf("GetBearerToken() err = %v, want %v", err, test.wantErr)
			}
		})
	}
}
