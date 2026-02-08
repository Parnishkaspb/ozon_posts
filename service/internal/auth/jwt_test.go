package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestToken_GenerateAndParse(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		login   string
		nameVal string
		surname string
	}{
		{
			name:    "ascii payload",
			secret:  "secret",
			login:   "Ivan",
			nameVal: "Иван",
			surname: "Грозный",
		},
		{
			name:    "unicode payload",
			secret:  "secret",
			login:   "Ivan",
			nameVal: "Иван",
			surname: "Грозный",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.secret, time.Minute)
			userID := uuid.New()

			token, err := svc.GenerateToken(userID, tt.login, tt.nameVal, tt.surname)
			if err != nil {
				t.Fatalf("generate token: %v", err)
			}

			u, err := svc.ParseToken(token)
			if err != nil {
				t.Fatalf("parse token: %v", err)
			}
			if u.ID != userID || u.Login != tt.login || u.Name != tt.nameVal || u.Surname != tt.surname {
				t.Fatalf("unexpected parsed user: %+v", u)
			}
		})
	}
}

func TestToken_ParseTokenErrors(t *testing.T) {
	svc := New("secret", time.Minute)
	good, err := svc.GenerateToken(uuid.New(), "Ivan", "Иван", "Грозный")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	tests := []struct {
		name  string
		svc   *Token
		token string
	}{
		{name: "malformed token", svc: svc, token: "not-a-jwt"},
		{name: "wrong secret", svc: New("another-secret", time.Minute), token: good},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.svc.ParseToken(tt.token); err == nil {
				t.Fatalf("expected parse error")
			}
		})
	}
}
