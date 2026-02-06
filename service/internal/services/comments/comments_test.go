package comments

import (
	"context"
	"errors"
	"github.com/Parnishkaspb/ozon_posts/internal/services/posts"
	"testing"

	"github.com/google/uuid"
)

type mockCommentRepo struct {
	called bool
}

func (m *mockCommentRepo) AnswerComment(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) error {
	m.called = true
	return nil
}

func (m *mockCommentRepo) CreateComment(ctx context.Context, text string, authorID, postID uuid.UUID) error {
	m.called = true
	return nil
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

	validText := "hello"
	longText := make([]byte, 2001)
	for i := range longText {
		longText[i] = 'a'
	}

	validAuthor := uuid.New()
	validPost := uuid.New()

	tests := []struct {
		name           string
		text           string
		authorID       uuid.UUID
		postID         uuid.UUID
		postAllows     bool
		wantErr        error
		wantRepoCalled bool
	}{
		{
			name:     "empty text",
			text:     "   ",
			authorID: validAuthor,
			postID:   validPost,
			wantErr:  posts.ErrTextRequired,
		},
		{
			name:     "text too long",
			text:     string(longText),
			authorID: validAuthor,
			postID:   validPost,
			wantErr:  ErrMax2000Symbols,
		},
		{
			name:     "empty authorID",
			text:     validText,
			authorID: uuid.Nil,
			postID:   validPost,
			wantErr:  posts.ErrAuthorIDRequired,
		},
		{
			name:     "empty postID",
			text:     validText,
			authorID: validAuthor,
			postID:   uuid.Nil,
			wantErr:  ErrPostIDRequired,
		},
		{
			name:       "comments disabled for post",
			text:       validText,
			authorID:   validAuthor,
			postID:     validPost,
			postAllows: false,
			wantErr:    ErrCantWriteComment,
		},
		{
			name:           "success",
			text:           validText,
			authorID:       validAuthor,
			postID:         validPost,
			postAllows:     true,
			wantRepoCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			commentRepo := &mockCommentRepo{}
			postRepo := &mockPostRepo{
				withoutComment: tt.postAllows,
			}

			svc := &CommentService{
				commentRepo: commentRepo,
				postRepo:    postRepo,
			}

			err := svc.CommentCreate(ctx, tt.text, tt.authorID, tt.postID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if commentRepo.called != tt.wantRepoCalled {
				t.Fatalf("CreateComment called = %v, want %v", commentRepo.called, tt.wantRepoCalled)
			}
		})
	}
}

func TestCommentService_AnswerComment(t *testing.T) {
	ctx := context.Background()

	validText := "hello"
	longText := make([]byte, 2001)
	for i := range longText {
		longText[i] = 'a'
	}

	validAuthor := uuid.New()
	validPost := uuid.New()
	validParentID := uuid.New()

	tests := []struct {
		name           string
		text           string
		authorID       uuid.UUID
		postID         uuid.UUID
		parentID       uuid.UUID
		postAllows     bool
		wantErr        error
		wantRepoCalled bool
	}{
		{
			name:     "empty text",
			text:     "   ",
			authorID: validAuthor,
			postID:   validPost,
			parentID: validParentID,
			wantErr:  posts.ErrTextRequired,
		},
		{
			name:     "text too long",
			text:     string(longText),
			authorID: validAuthor,
			postID:   validPost,
			parentID: validParentID,
			wantErr:  ErrMax2000Symbols,
		},
		{
			name:     "empty authorID",
			text:     validText,
			authorID: uuid.Nil,
			postID:   validPost,
			parentID: validParentID,
			wantErr:  posts.ErrAuthorIDRequired,
		},
		{
			name:     "empty postID",
			text:     validText,
			authorID: validAuthor,
			postID:   uuid.Nil,
			parentID: validParentID,
			wantErr:  ErrPostIDRequired,
		},
		{
			name:     "empty parentID",
			text:     validText,
			authorID: validAuthor,
			postID:   validPost,
			parentID: uuid.Nil,
			wantErr:  ErrCommentIDRequired,
		},
		{
			name:           "success",
			text:           validText,
			authorID:       validAuthor,
			postID:         validPost,
			parentID:       validParentID,
			postAllows:     true,
			wantRepoCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			commentRepo := &mockCommentRepo{}
			postRepo := &mockPostRepo{
				withoutComment: tt.postAllows,
			}

			svc := &CommentService{
				commentRepo: commentRepo,
				postRepo:    postRepo,
			}

			err := svc.CommentAnswer(ctx, tt.text, tt.authorID, tt.postID, tt.parentID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if commentRepo.called != tt.wantRepoCalled {
				t.Fatalf("AnswerComment called = %v, want %v", commentRepo.called, tt.wantRepoCalled)
			}
		})
	}
}
