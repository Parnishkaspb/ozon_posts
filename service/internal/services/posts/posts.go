package posts

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
	ErrAuthorIDRequired = errors.New("authorID is required")
	ErrTextRequired     = errors.New("text is required")
	ErrProblemsWithIDs  = errors.New("problems with IDs")
)

const defaultPageSize = 20

type PostRepo interface {
	CreatePost(ctx context.Context, ownerID uuid.UUID, text string, withoutComment bool) (*models.Post, error)
	GetAllPosts(ctx context.Context) ([]*models.Post, error)
	GetPostsByID(ctx context.Context, id string) (*models.Post, error)
	WithoutComment(ctx context.Context, postID uuid.UUID) (bool, error)
	GetPostsPage(ctx context.Context, first int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Post, bool, error)
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

func (s *PostService) GetPostsByPage(ctx context.Context, first int, after string) ([]*servicepb.Post, string, bool, error) {

	if first <= 0 {
		first = defaultPageSize
	}

	var afterT *time.Time
	var afterID *uuid.UUID
	if after != "" {
		t, id, err := parsePostCursor(after)
		if err != nil {
			return nil, "", false, fmt.Errorf("bad cursor: %w", err)
		}
		afterT = &t
		afterID = &id
	}

	posts, hasNext, err := s.repo.GetPostsPage(ctx, first, afterT, afterID)
	if err != nil {
		return nil, "", false, err
	}

	out := make([]*servicepb.Post, 0, len(posts))
	for _, p := range posts {
		out = append(out, toPBPost(p))
	}

	endCursor := ""
	if len(posts) > 0 {
		last := posts[len(posts)-1]
		endCursor = makePostCursor(last.CreatedAt, last.ID)
	}

	return out, endCursor, hasNext, nil
}

func toPBPost(p *models.Post) *servicepb.Post {
	return &servicepb.Post{
		Id:             p.ID.String(),
		AuthorId:       p.AuthorID.String(),
		Text:           p.Text,
		WithoutComment: p.WithoutComment,
		CreatedAt:      timestamppb.New(p.CreatedAt),
		UpdatedAt:      timestamppb.New(p.UpdatedAt),
	}
}

func makePostCursor(createdAt time.Time, id uuid.UUID) string {
	raw := createdAt.UTC().Format(time.RFC3339Nano) + "|" + id.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func parsePostCursor(cur string) (time.Time, uuid.UUID, error) {
	b, err := base64.RawURLEncoding.DecodeString(cur)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor format")
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
