package auth

import (
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/helper"
	"net/http"
	"strings"
)

func AuthMiddleware(jwtService *Token, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		h := r.Header.Get("Authorization")
		const prefix = "Bearer "

		if strings.HasPrefix(h, prefix) {
			token := strings.TrimSpace(strings.TrimPrefix(h, prefix))
			if token != "" {
				if user, err := jwtService.ParseToken(token); err == nil {
					ctx := helper.WithUser(r.Context(), user)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
