package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/Parnishkaspb/ozon_posts/internal/app"
	"github.com/Parnishkaspb/ozon_posts/internal/auth"
	"github.com/Parnishkaspb/ozon_posts/internal/config"
	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newMemoryHandler(t *testing.T) *Handler {
	t.Helper()
	cfg := &config.Config{
		Storage: config.StorageConfig{Driver: "memory"},
		JWT:     config.Token{Secret: "secret", TTL: time.Minute},
	}
	a, err := app.New(context.Background(), cfg, auth.New("secret", time.Minute))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return New(a)
}

func TestHandler_AuthAndPostsFlow(t *testing.T) {
	h := newMemoryHandler(t)
	ctx := context.Background()

	loginResp, err := h.Login(ctx, &servicepb.LoginRequest{Login: "Ivan", Password: "MoscowNeverSleep"})
	if err != nil || loginResp.GetToken() == "" {
		t.Fatalf("login failed: %v", err)
	}

	usersResp, err := h.GetUsers(ctx, &servicepb.GetUsersRequest{})
	if err != nil || len(usersResp.GetUsers()) == 0 {
		t.Fatalf("get users failed: %v", err)
	}
	authorID := usersResp.GetUsers()[0].GetId()

	createResp, err := h.CreatePost(ctx, &servicepb.CreatePostRequest{AuthorId: authorID, Text: "hello", WithoutComment: true})
	if err != nil {
		t.Fatalf("create post failed: %v", err)
	}
	postID := createResp.GetPost().GetId()

	getResp, err := h.GetPost(ctx, &servicepb.GetPostRequest{Id: postID})
	if err != nil || getResp.GetPost() == nil {
		t.Fatalf("get post failed: %v", err)
	}

	pageResp, err := h.GetPosts(ctx, &servicepb.GetPostsRequest{First: 1})
	if err != nil {
		t.Fatalf("get posts failed: %v", err)
	}
	if len(pageResp.GetPosts()) != 1 || pageResp.GetPosts()[0].GetId() != postID {
		t.Fatalf("unexpected posts page")
	}
}

func TestHandler_CommentsHierarchyFlow(t *testing.T) {
	h := newMemoryHandler(t)
	ctx := context.Background()

	usersResp, err := h.GetUsers(ctx, &servicepb.GetUsersRequest{})
	if err != nil || len(usersResp.GetUsers()) == 0 {
		t.Fatalf("get users failed: %v", err)
	}
	authorID := usersResp.GetUsers()[0].GetId()

	postResp, err := h.CreatePost(ctx, &servicepb.CreatePostRequest{AuthorId: authorID, Text: "post", WithoutComment: true})
	if err != nil {
		t.Fatalf("create post failed: %v", err)
	}
	postID := postResp.GetPost().GetId()

	rootResp, err := h.CreateComment(ctx, &servicepb.CreateCommentRequest{PostId: postID, AuthorId: authorID, Text: "root"})
	if err != nil {
		t.Fatalf("create root comment failed: %v", err)
	}
	rootID := rootResp.GetComment().GetId()

	replyResp, err := h.CreateComment(ctx, &servicepb.CreateCommentRequest{PostId: postID, AuthorId: authorID, ParentId: rootID, Text: "reply"})
	if err != nil {
		t.Fatalf("create reply failed: %v", err)
	}
	if replyResp.GetComment().GetParentId() != rootID {
		t.Fatalf("unexpected parent id")
	}

	rootsPage, err := h.GetComments(ctx, &servicepb.GetCommentsRequest{PostId: postID, First: 10})
	if err != nil {
		t.Fatalf("get root comments failed: %v", err)
	}
	if len(rootsPage.GetComments()) != 1 || rootsPage.GetComments()[0].GetId() != rootID {
		t.Fatalf("unexpected root comments")
	}

	repliesPage, err := h.GetComments(ctx, &servicepb.GetCommentsRequest{PostId: postID, ParentId: rootID, First: 10})
	if err != nil {
		t.Fatalf("get replies failed: %v", err)
	}
	if len(repliesPage.GetComments()) != 1 || repliesPage.GetComments()[0].GetParentId() != rootID {
		t.Fatalf("unexpected replies")
	}
}

func TestHandler_CreatePostValidation(t *testing.T) {
	h := newMemoryHandler(t)
	ctx := context.Background()

	_, err := h.CreatePost(ctx, &servicepb.CreatePostRequest{AuthorId: "bad-uuid", Text: "x"})
	if err == nil {
		t.Fatalf("expected validation error")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}
