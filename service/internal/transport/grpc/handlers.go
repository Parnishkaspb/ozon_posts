package grpc

import (
	"context"
	"errors"
	"github.com/Parnishkaspb/ozon_posts/internal/app"
	"github.com/Parnishkaspb/ozon_posts/internal/auth"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/Parnishkaspb/ozon_posts/internal/services/comments"
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
	postsAnswer, endCursor, hasNext, err := h.app.PostSRV.GetPostsByPage(ctx, int(req.GetFirst()), req.GetAfter())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &servicepb.GetPostsResponse{
		Posts:       postsAnswer,
		EndCursor:   endCursor,
		HasNextPage: hasNext,
	}, nil
}

func (h *Handler) GetUsers(ctx context.Context, req *servicepb.GetUsersRequest) (*servicepb.GetUsersResponse, error) {
	users, err := h.app.UserSRV.GetUsersByIds(ctx, req.GetIds())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	result := make([]*servicepb.User, 0, len(users))
	for _, user := range users {
		result = append(result, &servicepb.User{
			Id:      user.ID.String(),
			Name:    user.Name,
			Surname: user.Surname,
		})
	}

	return &servicepb.GetUsersResponse{Users: result}, nil
}

func (h *Handler) CreateComment(ctx context.Context, req *servicepb.CreateCommentRequest) (*servicepb.CreateCommentResponse, error) {
	uuidPostId, err := uuid.Parse(req.GetPostId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "post_id must be a valid UUID")
	}

	err = h.app.PostSRV.CanWriteComment(ctx, uuidPostId)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	uuidAuthorId, err := uuid.Parse(req.GetAuthorId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "author_id must be a valid UUID")
	}

	var comment *models.Comment
	if req.GetParentId() != "" {
		uuidParentId, err := uuid.Parse(req.GetParentId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "parent_id must be a valid UUID")
		}

		comment, err = h.app.CommentSRV.CommentAnswer(ctx, req.GetText(), uuidAuthorId, uuidPostId, uuidParentId)
	} else {
		comment, err = h.app.CommentSRV.CommentCreate(ctx, req.GetText(), uuidAuthorId, uuidPostId)
	}

	if err != nil {
		if errors.Is(err, comments.ErrCantWriteComment) {
			return nil, status.Error(codes.InvalidArgument, comments.ErrCantWriteComment.Error())
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	parentID := ""
	if comment.ParentCommentID != nil {
		parentID = comment.ParentCommentID.String()
	}

	return &servicepb.CreateCommentResponse{Comment: &servicepb.Comment{
		Id:        comment.ID.String(),
		PostId:    comment.PostID.String(),
		AuthorId:  comment.AuthorID.String(),
		ParentId:  parentID,
		Text:      comment.Text,
		CreatedAt: timestamppb.New(comment.CreatedAt),
	}}, nil
}
