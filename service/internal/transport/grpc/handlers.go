package grpc

import (
	"context"
	"errors"
	"github.com/Parnishkaspb/ozon_posts/internal/app"
	"github.com/Parnishkaspb/ozon_posts/internal/auth"
	"github.com/Parnishkaspb/ozon_posts/internal/services/posts"
	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	servicepb.UnimplementedAuthServiceServer
	servicepb.UnimplementedUserServiceServer
	servicepb.UnimplementedPostServiceServer
	servicepb.UnimplementedCommentServiceServer

	app *app.App
}

func New(a *app.App) *Handler {
	return &Handler{app: a}
}

func (h *Handler) Login(
	ctx context.Context,
	req *servicepb.LoginRequest,
) (*servicepb.LoginResponse, error) {

	token, err := h.app.Auth.Authenticate(ctx, req.GetLogin(), req.GetPassword())

	if err != nil {
		if errors.Is(err, auth.ErrEmpty) {
			return nil, status.Error(codes.InvalidArgument, auth.ErrEmpty.Error())
		}
		if errors.Is(err, auth.ErrIncorrect) {
			return nil, status.Error(codes.Unauthenticated, "invalid login or password")
		}
		return nil, status.Error(codes.Internal, "internal authentication error")
	}

	return &servicepb.LoginResponse{
		Token: token,
	}, nil
}

func (h *Handler) CreatePost(
	ctx context.Context,
	req *servicepb.CreatePostRequest,
) (*servicepb.CreatePostResponse, error) {

	authorID, err := uuid.Parse(req.GetAuthorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "author_id must be a valid UUID")
	}

	post, err := h.app.PostSRV.CreatePost(ctx, authorID, req.GetText(), req.GetWithoutComment())

	if err != nil {
		switch {
		case errors.Is(err, posts.ErrAuthorIDRequired):
			return nil, status.Error(codes.InvalidArgument, posts.ErrAuthorIDRequired.Error())
		case errors.Is(err, posts.ErrTextRequired):
			return nil, status.Error(codes.InvalidArgument, posts.ErrTextRequired.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &servicepb.CreatePostResponse{
		Post: &servicepb.Post{
			Id:             post.ID.String(),
			AuthorId:       post.AuthorID.String(),
			Text:           post.Text,
			WithoutComment: post.WithoutComment,
			CreatedAt:      timestamppb.New(post.CreatedAt),
			UpdatedAt:      timestamppb.New(post.UpdatedAt),
		},
	}, nil
}

func (h *Handler) GetPosts(ctx context.Context, req *servicepb.GetPostsRequest) (*servicepb.GetPostsResponse, error) {
	postsAnswer, err := h.app.PostSRV.GetAllPosts(ctx, req.Ids)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	result := make([]*servicepb.Post, 0, len(postsAnswer))
	for _, post := range postsAnswer {
		result = append(result, &servicepb.Post{
			Id:             post.ID.String(),
			AuthorId:       post.AuthorID.String(),
			Text:           post.Text,
			WithoutComment: post.WithoutComment,
			CreatedAt:      timestamppb.New(post.CreatedAt),
			UpdatedAt:      timestamppb.New(post.UpdatedAt),
		})
	}

	return &servicepb.GetPostsResponse{Posts: result}, nil
}
