package auth

import (
	"github.com/Parnishkaspb/ozon_posts/internal/models"
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

func (t *Token) GenerateToken(userID uuid.UUID, login, name, surname string) (string, error) {
	claims := Claims{
		UserID:  userID,
		Login:   login,
		Name:    name,
		Surname: surname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.Secret))
}

func (t *Token) ParseToken(tokenString string) (*models.User, error) {
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

	return &models.User{
		ID:      claims.UserID,
		Login:   claims.Login,
		Name:    claims.Name,
		Surname: claims.Surname,
	}, nil
}
