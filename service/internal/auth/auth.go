package auth

import (
	"context"
	"errors"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/Parnishkaspb/ozon_posts/internal/repositories/users"
	"github.com/google/uuid"
)

type ContextKey string

const UserContextKey ContextKey = "auth_user"

var (
	ErrEmpty     = errors.New("login or password is empty")
	ErrIncorrect = errors.New("login or password is incorrect")
)

type JWT interface {
	GenerateToken(userID uuid.UUID, login, name, surname string) (string, error)
}

type UserRepo interface {
	GetUserByLoginPassword(ctx context.Context, login, password string) (*models.User, error)
}

type Auth struct {
	jwtService JWT
	userRepo   UserRepo
}

func NewAuth(jwt JWT, userRepo UserRepo) *Auth {
	return &Auth{jwtService: jwt, userRepo: userRepo}
}

func (a *Auth) Authenticate(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", ErrEmpty
	}

	user, err := a.userRepo.GetUserByLoginPassword(ctx, login, password)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return "", ErrIncorrect
		}
		return "", err
	}

	return a.jwtService.GenerateToken(user.ID, login, user.Name, user.Surname)
}
