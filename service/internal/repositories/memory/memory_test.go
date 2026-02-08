package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestUserRepo_GetUserByLoginPassword(t *testing.T) {
	repo := NewUserRepo(NewStore())
	ctx := context.Background()

	tests := []struct {
		name     string
		login    string
		password string
		wantErr  bool
	}{
		{
			name:     "valid seeded user",
			login:    "Ivan",
			password: "MoscowNeverSleep",
		},
		{
			name:     "wrong password",
			login:    "Ivan",
			password: "MoscowEveryNightSleep",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := repo.GetUserByLoginPassword(ctx, tt.login, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u == nil || u.Login != "Ivan" {
				t.Fatalf("expected seeded user")
			}
		})
	}
}

func TestPostRepo_CreateAndPage(t *testing.T) {
	store := NewStore()
	repo := NewPostRepo(store)
	ctx := context.Background()
	author := uuid.New()

	first, err := repo.CreatePost(ctx, author, "first", false)
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	second, err := repo.CreatePost(ctx, author, "second", false)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	after := second.CreatedAt
	afterID := second.ID
	tests := []struct {
		name         string
		first        int
		afterCreated *time.Time
		afterID      *uuid.UUID
		wantLen      int
		wantHasNext  bool
		wantFirstID  uuid.UUID
	}{
		{name: "first page", first: 1, wantLen: 1, wantHasNext: true, wantFirstID: second.ID},
		{name: "second page", first: 1, afterCreated: &after, afterID: &afterID, wantLen: 1, wantHasNext: false, wantFirstID: first.ID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, hasNext, err := repo.GetPostsPage(ctx, tt.first, tt.afterCreated, tt.afterID)
			if err != nil {
				t.Fatalf("get page: %v", err)
			}
			if hasNext != tt.wantHasNext || len(page) != tt.wantLen {
				t.Fatalf("unexpected page meta")
			}
			if len(page) > 0 && page[0].ID != tt.wantFirstID {
				t.Fatalf("unexpected first item")
			}
		})
	}
}

func TestPostRepo_WithoutCommentNotFound(t *testing.T) {
	repo := NewPostRepo(NewStore())
	_, err := repo.WithoutComment(context.Background(), uuid.New())
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestCommentRepo_CreateAndReplies(t *testing.T) {
	repo := NewCommentRepo(NewStore())
	ctx := context.Background()
	postID := uuid.New()
	author := uuid.New()

	root, err := repo.CreateComment(ctx, "root", author, postID)
	if err != nil {
		t.Fatalf("create root: %v", err)
	}
	reply, err := repo.AnswerComment(ctx, "reply", author, postID, root.ID)
	if err != nil {
		t.Fatalf("answer root: %v", err)
	}

	tests := []struct {
		name     string
		parentID *uuid.UUID
		wantID   uuid.UUID
	}{
		{name: "roots", parentID: nil, wantID: root.ID},
		{name: "replies", parentID: &root.ID, wantID: reply.ID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := repo.GetCommentsPage(ctx, postID, tt.parentID, 10, nil, nil)
			if err != nil {
				t.Fatalf("get comments: %v", err)
			}
			if len(items) != 1 || items[0].ID != tt.wantID {
				t.Fatalf("unexpected items")
			}
		})
	}

	if reply.ParentCommentID == nil || *reply.ParentCommentID != root.ID {
		t.Fatalf("reply parent mismatch")
	}
}
