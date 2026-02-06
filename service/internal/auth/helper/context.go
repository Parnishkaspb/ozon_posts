package helper

import (
	"context"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
)

type contextKey string

const userContextKey contextKey = "auth_user"

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func FromContext(ctx context.Context) (*models.User, bool) {

	u, ok := ctx.Value(userContextKey).(*models.User)
	return u, ok
}
