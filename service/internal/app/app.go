package app

import (
	"context"
	"fmt"
	"github.com/Parnishkaspb/ozon_posts/internal/auth"
	"github.com/Parnishkaspb/ozon_posts/internal/database/postgresql"
	commentrepo "github.com/Parnishkaspb/ozon_posts/internal/repositories/comments"
	postrepo "github.com/Parnishkaspb/ozon_posts/internal/repositories/posts"
	userrepo "github.com/Parnishkaspb/ozon_posts/internal/repositories/users"
	commentsrv "github.com/Parnishkaspb/ozon_posts/internal/services/comments"
	postsrv "github.com/Parnishkaspb/ozon_posts/internal/services/posts"
	usersrv "github.com/Parnishkaspb/ozon_posts/internal/services/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Pool       *pgxpool.Pool
	PostSRV    *postsrv.PostService
	CommentSRV *commentsrv.CommentService
	UserSRV    *usersrv.UserService
	Auth       *auth.Auth
}

func New(ctx context.Context, dsn string, jwtService *auth.Token) (*App, error) {
	db, err := postgresql.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgresql.New: %w", err)
	}

	pool := db.Pool

	userRepo := userrepo.New(pool)
	postRepo := postrepo.New(pool)
	commentRepo := commentrepo.New(pool)

	authService := auth.NewAuth(jwtService, userRepo)
	postService := postsrv.New(postRepo)
	userService := usersrv.NewUserService(userRepo)
	commentService := commentsrv.New(commentRepo, postRepo)

	return &App{
		Pool:       pool,
		PostSRV:    postService,
		CommentSRV: commentService,
		UserSRV:    userService,
		Auth:       authService,
	}, nil
}

func (a *App) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
}
