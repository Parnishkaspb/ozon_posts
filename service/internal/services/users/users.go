package users

import (
	"context"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
)

type UserRepo interface {
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type UserService struct {
	repo UserRepo
}

func NewUserService(repo UserRepo) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAllUsers(ctx)
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}
