package posts

import (
	"context"
	"errors"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrAuthorIDRequired = errors.New("authorID is required")
	ErrTextRequired     = errors.New("text is required")
	ErrProblemsWithIDs  = errors.New("problems with IDs")
)

type PostRepo interface {
	CreatePost(ctx context.Context, ownerID uuid.UUID, text string, withoutComment bool) (*models.Post, error)
	GetAllPosts(ctx context.Context) ([]*models.Post, error)
	GetPostsByID(ctx context.Context, id string) (*models.Post, error)
	WithoutComment(ctx context.Context, postID uuid.UUID) (bool, error)
}

type PostService struct {
	repo PostRepo
}

func New(repo PostRepo) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(ctx context.Context, authorID uuid.UUID, text string, withoutComment bool) (*models.Post, error) {
	if authorID == uuid.Nil {
		return nil, ErrAuthorIDRequired
	}

	if strings.TrimSpace(text) == "" {
		return nil, ErrTextRequired
	}

	return s.repo.CreatePost(ctx, authorID, text, withoutComment)
}

func (s *PostService) GetAllPosts(ctx context.Context, id []string) ([]*models.Post, error) {
	if len(id) > 1 {
		return nil, ErrProblemsWithIDs
	}

	if len(id) == 1 && id[0] != "" {
		post, err := s.repo.GetPostsByID(ctx, id[0])
		if err != nil {
			return nil, err
		}
		if post == nil {
			return []*models.Post{}, nil
		}
		return []*models.Post{post}, nil
	}

	return s.repo.GetAllPosts(ctx)
}

func (s *PostService) CanWriteComment(ctx context.Context, id uuid.UUID) error {
	ok, err := s.repo.WithoutComment(ctx, id)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("can't write a comment to this post")
	}

	return nil
}
