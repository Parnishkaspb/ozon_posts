package comments

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
)

type mockCommentRepo struct {
	createCalled bool
	answerCalled bool
	createFn     func(ctx context.Context, text string, authorID, postID uuid.UUID) (*models.Comment, error)
	answerFn     func(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) (*models.Comment, error)
}

func (m *mockCommentRepo) CreateComment(ctx context.Context, text string, authorID, postID uuid.UUID) (*models.Comment, error) {
	m.createCalled = true
	if m.createFn != nil {
		return m.createFn(ctx, text, authorID, postID)
	}
	return &models.Comment{}, nil
}

func (m *mockCommentRepo) AnswerComment(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) (*models.Comment, error) {
	m.answerCalled = true
	if m.answerFn != nil {
		return m.answerFn(ctx, text, authorID, postID, commentID)
	}
	return &models.Comment{}, nil
}

func (m *mockCommentRepo) GetCommentsPage(ctx context.Context, postID uuid.UUID, parentID *uuid.UUID, limit int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Comment, error) {
	return nil, nil
}

type mockPostRepo struct {
	withoutComment bool
	err            error
}

func (m *mockPostRepo) WithoutComment(ctx context.Context, postID uuid.UUID) (bool, error) {
	return m.withoutComment, m.err
}

func TestCommentService_CommentCreate(t *testing.T) {
	ctx := context.Background()
	validAuthor := uuid.New()
	validPost := uuid.New()

	t.Run("text required", func(t *testing.T) {
		svc := New(&mockCommentRepo{}, &mockPostRepo{})
		_, err := svc.CommentCreate(ctx, "   ", validAuthor, validPost)
		if !errors.Is(err, ErrTextRequired) {
			t.Fatalf("expected %v, got %v", ErrTextRequired, err)
		}
	})

	t.Run("text too long", func(t *testing.T) {
		svc := New(&mockCommentRepo{}, &mockPostRepo{})
		_, err := svc.CommentCreate(ctx, strings.Repeat("a", 2001), validAuthor, validPost)
		if !errors.Is(err, ErrMax2000Symbols) {
			t.Fatalf("expected %v, got %v", ErrMax2000Symbols, err)
		}
	})

	t.Run("author required", func(t *testing.T) {
		svc := New(&mockCommentRepo{}, &mockPostRepo{})
		_, err := svc.CommentCreate(ctx, "ok", uuid.Nil, validPost)
		if !errors.Is(err, ErrAuthorIDRequired) {
			t.Fatalf("expected %v, got %v", ErrAuthorIDRequired, err)
		}
	})

	t.Run("post required", func(t *testing.T) {
		svc := New(&mockCommentRepo{}, &mockPostRepo{})
		_, err := svc.CommentCreate(ctx, "ok", validAuthor, uuid.Nil)
		if !errors.Is(err, ErrPostIDRequired) {
			t.Fatalf("expected %v, got %v", ErrPostIDRequired, err)
		}
	})

	t.Run("comments disabled", func(t *testing.T) {
		commentRepo := &mockCommentRepo{}
		svc := New(commentRepo, &mockPostRepo{withoutComment: false})
		_, err := svc.CommentCreate(ctx, "ok", validAuthor, validPost)
		if !errors.Is(err, ErrCantWriteComment) {
			t.Fatalf("expected %v, got %v", ErrCantWriteComment, err)
		}
		if commentRepo.createCalled {
			t.Fatalf("comment repo must not be called")
		}
	})

	t.Run("success", func(t *testing.T) {
		commentRepo := &mockCommentRepo{}
		svc := New(commentRepo, &mockPostRepo{withoutComment: true})
		_, err := svc.CommentCreate(ctx, "ok", validAuthor, validPost)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !commentRepo.createCalled {
			t.Fatalf("create repo must be called")
		}
	})
}

func TestCommentService_CommentAnswer(t *testing.T) {
	ctx := context.Background()
	validAuthor := uuid.New()
	validPost := uuid.New()
	validParent := uuid.New()

	t.Run("comment id required", func(t *testing.T) {
		svc := New(&mockCommentRepo{}, &mockPostRepo{})
		_, err := svc.CommentAnswer(ctx, "ok", validAuthor, validPost, uuid.Nil)
		if !errors.Is(err, ErrCommentIDRequired) {
			t.Fatalf("expected %v, got %v", ErrCommentIDRequired, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		commentRepo := &mockCommentRepo{}
		svc := New(commentRepo, &mockPostRepo{})
		_, err := svc.CommentAnswer(ctx, "ok", validAuthor, validPost, validParent)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !commentRepo.answerCalled {
			t.Fatalf("answer repo must be called")
		}
	})
}
