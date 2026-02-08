package comments

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPostIDRequired    = errors.New("postID is required")
	ErrMax2000Symbols    = errors.New("text max 2000 symbols")
	ErrCommentIDRequired = errors.New("commentID is required")
	ErrCantWriteComment  = errors.New("can't write a comment to this post")
	ErrAuthorIDRequired  = errors.New("authorID is required")
	ErrTextRequired      = errors.New("text is required")
	ErrInvalidCursor     = errors.New("invalid cursor")
	ErrInvalidParentID   = errors.New("parentID is invalid")
	ErrBadFirst          = errors.New("first must be > 0")
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type CommentRepo interface {
	CreateComment(ctx context.Context, text string, authorID, postID uuid.UUID) (*models.Comment, error)
	AnswerComment(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) (*models.Comment, error)
	GetCommentsPage(ctx context.Context, postID uuid.UUID, parentID *uuid.UUID, limit int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Comment, error)
}

type PostRepo interface {
	WithoutComment(ctx context.Context, postID uuid.UUID) (bool, error)
}

type CommentService struct {
	commentRepo CommentRepo
	postRepo    PostRepo
}

func New(comment CommentRepo, post PostRepo) *CommentService {
	return &CommentService{
		commentRepo: comment,
		postRepo:    post,
	}
}

func (s *CommentService) normalizeAndValidateText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ErrTextRequired
	}
	if len([]rune(text)) > 2000 {
		return "", ErrMax2000Symbols
	}
	return text, nil
}

func (s *CommentService) requireUUID(id uuid.UUID, err error) error {
	if id == uuid.Nil {
		return err
	}
	return nil
}

func (s *CommentService) CommentCreate(ctx context.Context, text string, authorID, postID uuid.UUID) (*models.Comment, error) {
	text, err := s.normalizeAndValidateText(text)
	if err != nil {
		return &models.Comment{}, err
	}
	if err := s.requireUUID(authorID, ErrAuthorIDRequired); err != nil {
		return &models.Comment{}, err
	}
	if err := s.requireUUID(postID, ErrPostIDRequired); err != nil {
		return &models.Comment{}, err
	}

	ok, err := s.postRepo.WithoutComment(ctx, postID)
	if err != nil {
		return &models.Comment{}, err
	}

	if !ok {
		return &models.Comment{}, ErrCantWriteComment
	}

	return s.commentRepo.CreateComment(ctx, text, authorID, postID)
}

func (s *CommentService) CommentAnswer(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) (*models.Comment, error) {
	text, err := s.normalizeAndValidateText(text)
	if err != nil {
		return &models.Comment{}, err
	}
	if err := s.requireUUID(authorID, ErrAuthorIDRequired); err != nil {
		return &models.Comment{}, err
	}
	if err := s.requireUUID(postID, ErrPostIDRequired); err != nil {
		return &models.Comment{}, err
	}
	if err := s.requireUUID(commentID, ErrCommentIDRequired); err != nil {
		return &models.Comment{}, err
	}

	return s.commentRepo.AnswerComment(ctx, text, authorID, postID, commentID)
}

func (s *CommentService) GetComments(ctx context.Context, req *servicepb.GetCommentsRequest) (*servicepb.GetCommentsResponse, error) {
	if req.GetPostId() == "" {
		return nil, ErrPostIDRequired
	}
	postID, err := uuid.Parse(req.GetPostId())
	if err != nil {
		return nil, ErrPostIDRequired
	}

	// parent_id: "" => root, иначе uuid
	var parentID *uuid.UUID
	if req.GetParentId() != "" {
		p, err := uuid.Parse(req.GetParentId())
		if err != nil {
			return nil, ErrInvalidParentID
		}
		parentID = &p
	}

	first := int(req.GetFirst())
	if first == 0 {
		first = defaultPageSize
	}
	if first < 0 {
		return nil, ErrBadFirst
	}
	if first > maxPageSize {
		first = maxPageSize
	}
	limit := first + 1

	// cursor
	var afterCreatedAt *time.Time
	var afterID *uuid.UUID
	if req.GetAfter() != "" {
		t, id, err := parseCursor(req.GetAfter())
		if err != nil {
			return nil, ErrInvalidCursor
		}
		afterCreatedAt = &t
		afterID = &id
	}

	items, err := s.commentRepo.GetCommentsPage(ctx, postID, parentID, limit, afterCreatedAt, afterID)
	if err != nil {
		return nil, err
	}

	hasNext := false
	if len(items) > first {
		hasNext = true
		items = items[:first]
	}

	endCursor := ""
	if len(items) > 0 {
		last := items[len(items)-1]
		endCursor = makeCursor(last.CreatedAt, last.ID)
	}

	out := make([]*servicepb.Comment, 0, len(items))
	for _, c := range items {
		out = append(out, toPB(c))
	}

	return &servicepb.GetCommentsResponse{
		Comments:    out,
		EndCursor:   endCursor,
		HasNextPage: hasNext,
	}, nil
}

func toPB(c *models.Comment) *servicepb.Comment {
	parent := ""
	if c.ParentCommentID != nil {
		parent = c.ParentCommentID.String()
	}

	return &servicepb.Comment{
		Id:        c.ID.String(),
		PostId:    c.PostID.String(),
		AuthorId:  c.AuthorID.String(),
		ParentId:  parent, // "" для root
		Text:      c.Text,
		CreatedAt: timestamppb.New(c.CreatedAt),
	}
}

func makeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := createdAt.UTC().Format(time.RFC3339Nano) + "|" + id.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func parseCursor(cur string) (time.Time, uuid.UUID, error) {
	b, err := base64.RawURLEncoding.DecodeString(cur)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("bad cursor format")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	return t, id, nil
}
