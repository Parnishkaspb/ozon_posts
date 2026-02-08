package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/Parnishkaspb/ozon_posts/internal/auth"
	"github.com/Parnishkaspb/ozon_posts/internal/config"
	"github.com/Parnishkaspb/ozon_posts/internal/database/postgresql"
	commentrepo "github.com/Parnishkaspb/ozon_posts/internal/repositories/comments"
	"github.com/Parnishkaspb/ozon_posts/internal/repositories/memory"
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

var ErrUnknownStorageDriver = errors.New("unknown storage driver")

type userRepository interface {
	usersrv.UserRepo
	auth.UserRepo
}

func New(ctx context.Context, cfg *config.Config, jwtService *auth.Token) (*App, error) {
	driver := cfg.Storage.Driver
	if driver == "" {
		driver = "postgres"
	}

	var (
		pool        *pgxpool.Pool
		userRepo    userRepository
		postRepo    postsrv.PostRepo
		commentRepo commentsrv.CommentRepo
	)

	switch driver {
	case "postgres":
		db, err := postgresql.New(ctx, cfg.PostgresDSN())
		if err != nil {
			return nil, fmt.Errorf("postgresql.New: %w", err)
		}
		pool = db.Pool
		userRepo = userrepo.New(pool)
		postRepo = postrepo.New(pool)
		commentRepo = commentrepo.New(pool)
	case "memory":
		store := memory.NewStore()
		userRepo = memory.NewUserRepo(store)
		postRepo = memory.NewPostRepo(store)
		commentRepo = memory.NewCommentRepo(store)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownStorageDriver, driver)
	}

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
