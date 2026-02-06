package auth

import (
	helper "github.com/Parnishkaspb/ozon_posts_graphql/internal/helper/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type Token struct {
	Secret string
	TTL    time.Duration
}

type Claims struct {
	UserID  uuid.UUID `json:"user_id"`
	Login   string    `json:"login"`
	Name    string    `json:"name"`
	Surname string    `json:"surname"`
	jwt.RegisteredClaims
}

func New(secret string, ttl time.Duration) *Token {
	return &Token{
		Secret: secret,
		TTL:    ttl,
	}
}

func (t *Token) ParseToken(tokenString string) (*helper.User, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(t.Secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return &helper.User{
		ID:      claims.UserID,
		Login:   claims.Login,
		Name:    claims.Name,
		Surname: claims.Surname,
	}, nil
}
