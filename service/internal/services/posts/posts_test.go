package posts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
)

type mockPostRepo struct {
	createFn func(ctx context.Context, ownerID uuid.UUID, text string, withoutComment bool) (*models.Post, error)
}

func (m *mockPostRepo) CreatePost(ctx context.Context, ownerID uuid.UUID, text string, withoutComment bool) (*models.Post, error) {
	if m.createFn != nil {
		return m.createFn(ctx, ownerID, text, withoutComment)
	}
	return nil, nil
}

func (m *mockPostRepo) GetAllPosts(ctx context.Context) ([]*models.Post, error) {
	return nil, nil
}

func (m *mockPostRepo) GetPostsByID(ctx context.Context, id string) (*models.Post, error) {
	return nil, nil
}

func (m *mockPostRepo) WithoutComment(ctx context.Context, postID uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockPostRepo) GetPostsPage(ctx context.Context, first int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Post, bool, error) {
	return nil, false, nil
}

func TestPostService_CreatePost(t *testing.T) {
	ctx := context.Background()
	authorID := uuid.New()
	repoErr := errors.New("repo failed")

	t.Run("author id required", func(t *testing.T) {
		svc := New(&mockPostRepo{})
		_, err := svc.CreatePost(ctx, uuid.Nil, "text", false)
		if !errors.Is(err, ErrAuthorIDRequired) {
			t.Fatalf("expected %v, got %v", ErrAuthorIDRequired, err)
		}
	})

	t.Run("text required", func(t *testing.T) {
		svc := New(&mockPostRepo{})
		_, err := svc.CreatePost(ctx, authorID, "   ", false)
		if !errors.Is(err, ErrTextRequired) {
			t.Fatalf("expected %v, got %v", ErrTextRequired, err)
		}
	})

	t.Run("repo error returned", func(t *testing.T) {
		svc := New(&mockPostRepo{createFn: func(ctx context.Context, ownerID uuid.UUID, text string, withoutComment bool) (*models.Post, error) {
			return nil, repoErr
		}})
		_, err := svc.CreatePost(ctx, authorID, "ok", true)
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		var (
			gotAuthor  uuid.UUID
			gotText    string
			gotWithout bool
		)

		expected := &models.Post{ID: uuid.New(), AuthorID: authorID, Text: "ok", WithoutComment: true}
		svc := New(&mockPostRepo{createFn: func(ctx context.Context, ownerID uuid.UUID, text string, withoutComment bool) (*models.Post, error) {
			gotAuthor = ownerID
			gotText = text
			gotWithout = withoutComment
			return expected, nil
		}})

		got, err := svc.CreatePost(ctx, authorID, "ok", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != expected {
			t.Fatalf("expected same post pointer")
		}
		if gotAuthor != authorID || gotText != "ok" || !gotWithout {
			t.Fatalf("unexpected repo args: author=%s text=%q without=%v", gotAuthor, gotText, gotWithout)
		}
	})
}
