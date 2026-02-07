package users

import (
	"context"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
	"log"
)

type UserRepo interface {
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*models.User, error)
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

func (s *UserService) GetUsersByIds(ctx context.Context, ids []string) ([]*models.User, error) {
	log.Printf("IDS: +%v", ids)
	if len(ids) == 0 {
		return s.repo.GetAllUsers(ctx)
	}
	return s.repo.GetUsersByIDs(ctx, ids)
}
