package helper

import (
	"context"
	helper "github.com/Parnishkaspb/ozon_posts_graphql/internal/helper/model"
)

type contextKey string

const userContextKey contextKey = "auth_user"

func WithUser(ctx context.Context, user *helper.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func FromContext(ctx context.Context) (*helper.User, bool) {

	u, ok := ctx.Value(userContextKey).(*helper.User)
	return u, ok
}
