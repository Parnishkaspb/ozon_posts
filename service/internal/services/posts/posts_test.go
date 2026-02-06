package posts

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
)

type mockRepo struct {
	gotauthorID       uuid.UUID
	gotText           string
	gotWithoutComment bool

	retPost *models.Post
	retErr  error

	called int
}

func (m *mockRepo) CreatePost(ctx context.Context, authorID uuid.UUID, text string, withoutComment bool) (*models.Post, error) {
	m.called++
	m.gotauthorID = authorID
	m.gotText = text
	m.gotWithoutComment = withoutComment
	return m.retPost, m.retErr
}

func boolPtr(v bool) *bool { return &v }

func TestPostService_CreatePost(t *testing.T) {
	ctx := context.Background()
	validAuthorID := uuid.New()

	okPost := &models.Post{
		ID:             uuid.New(),
		AuthorID:       validAuthorID,
		Text:           "hello",
		WithoutComment: true,
	}

	repoErr := errors.New("repo failed")

	tests := []struct {
		name           string
		authorID       uuid.UUID
		text           string
		inWithoutPtr   *bool
		wantWithout    bool
		repoReturnPost *models.Post
		repoReturnErr  error
		wantErr        bool
		wantCalls      int
		wantErrIs      error
	}{
		{
			name:           "nil withoutComment -> default false",
			authorID:       validAuthorID,
			text:           "a",
			inWithoutPtr:   nil,
			wantWithout:    false,
			repoReturnPost: okPost,
			wantErr:        false,
			wantCalls:      1,
		},
		{
			name:           "withoutComment true",
			authorID:       validAuthorID,
			text:           "b",
			inWithoutPtr:   boolPtr(true),
			wantWithout:    true,
			repoReturnPost: okPost,
			wantErr:        false,
			wantCalls:      1,
		},
		{
			name:           "withoutComment false",
			authorID:       validAuthorID,
			text:           "c",
			inWithoutPtr:   boolPtr(false),
			wantWithout:    false,
			repoReturnPost: okPost,
			wantErr:        false,
			wantCalls:      1,
		},
		{
			name:         "authorID required",
			authorID:     uuid.Nil,
			text:         "text",
			inWithoutPtr: nil,
			wantErr:      true,
			wantCalls:    0,
			wantErrIs:    ErrAuthorIDRequired,
		},
		{
			name:         "text required empty",
			authorID:     validAuthorID,
			text:         "",
			inWithoutPtr: nil,
			wantErr:      true,
			wantCalls:    0,
			wantErrIs:    ErrTextRequired,
		},
		{
			name:         "text required spaces",
			authorID:     validAuthorID,
			text:         "   ",
			inWithoutPtr: nil,
			wantErr:      true,
			wantCalls:    0,
			wantErrIs:    ErrTextRequired,
		},
		{
			name:          "repo error is returned",
			authorID:      validAuthorID,
			text:          "d",
			inWithoutPtr:  boolPtr(true),
			wantWithout:   true,
			repoReturnErr: repoErr,
			wantErr:       true,
			wantCalls:     1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := &mockRepo{
				retPost: tt.repoReturnPost,
				retErr:  tt.repoReturnErr,
			}
			svc := New(m)

			got, err := svc.CreatePost(ctx, tt.authorID, tt.text, tt.inWithoutPtr)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if m.called != tt.wantCalls {
				t.Fatalf("expected repo called %d times, got %d", tt.wantCalls, m.called)
			}

			if tt.wantCalls == 1 {
				if m.gotauthorID != tt.authorID {
					t.Fatalf("authorID: expected %s, got %s", tt.authorID, m.gotauthorID)
				}
				if m.gotText != tt.text {
					t.Fatalf("text: expected %q, got %q", tt.text, m.gotText)
				}
				if m.gotWithoutComment != tt.wantWithout {
					t.Fatalf("withoutComment: expected %v, got %v", tt.wantWithout, m.gotWithoutComment)
				}
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.repoReturnPost) {
				t.Fatalf("post: expected %+v, got %+v", tt.repoReturnPost, got)
			}
		})
	}
}
