package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	userrepo "github.com/Parnishkaspb/ozon_posts/internal/repositories/users"
	"github.com/google/uuid"
)

type mockJWT struct {
	token      string
	err        error
	gotUserID  uuid.UUID
	gotLogin   string
	gotName    string
	gotSurname string
}

func (m *mockJWT) GenerateToken(userID uuid.UUID, login, name, surname string) (string, error) {
	m.gotUserID = userID
	m.gotLogin = login
	m.gotName = name
	m.gotSurname = surname
	return m.token, m.err
}

type mockUserRepo struct {
	user *models.User
	err  error
}

func (m *mockUserRepo) GetUserByLoginPassword(ctx context.Context, login, password string) (*models.User, error) {
	return m.user, m.err
}

func TestAuth_Authenticate(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")
	jwtErr := errors.New("sign failed")
	u := &models.User{ID: uuid.New(), Name: "Ivan", Surname: "Grozniy"}

	tests := []struct {
		name        string
		login       string
		password    string
		repo        *mockUserRepo
		jwt         *mockJWT
		wantErr     error
		wantToken   string
		checkJWTArg bool
	}{
		{
			name:    "empty credentials",
			wantErr: ErrEmpty,
			repo:    &mockUserRepo{},
			jwt:     &mockJWT{},
		},
		{
			name:     "user not found mapped to incorrect",
			login:    "login",
			password: "password",
			repo:     &mockUserRepo{err: userrepo.ErrUserNotFound},
			jwt:      &mockJWT{},
			wantErr:  ErrIncorrect,
		},
		{
			name:     "repo error returned",
			login:    "login",
			password: "password",
			repo:     &mockUserRepo{err: repoErr},
			jwt:      &mockJWT{},
			wantErr:  repoErr,
		},
		{
			name:        "jwt error returned",
			login:       "login",
			password:    "password",
			repo:        &mockUserRepo{user: u},
			jwt:         &mockJWT{err: jwtErr},
			wantErr:     jwtErr,
			checkJWTArg: true,
		},
		{
			name:        "success",
			login:       "login",
			password:    "password",
			repo:        &mockUserRepo{user: u},
			jwt:         &mockJWT{token: "token"},
			wantToken:   "token",
			checkJWTArg: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAuth(tt.jwt, tt.repo)
			tok, err := a.Authenticate(ctx, tt.login, tt.password)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tok != tt.wantToken {
				t.Fatalf("expected token %q, got %q", tt.wantToken, tok)
			}
			if tt.checkJWTArg {
				if tt.jwt.gotUserID != u.ID || tt.jwt.gotLogin != tt.login || tt.jwt.gotName != u.Name || tt.jwt.gotSurname != u.Surname {
					t.Fatalf("unexpected jwt args")
				}
			}
		})
	}
}
