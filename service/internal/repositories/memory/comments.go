package memory

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
)

var ErrParentNotFound = errors.New("parent comment not found")

type CommentRepo struct {
	store *Store
}

func NewCommentRepo(store *Store) *CommentRepo {
	return &CommentRepo{store: store}
}

func (r *CommentRepo) CreateComment(ctx context.Context, text string, authorID, postID uuid.UUID) (*models.Comment, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	c := &models.Comment{
		ID:        uuid.New(),
		PostID:    postID,
		AuthorID:  authorID,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}
	r.store.comments[c.ID] = copyComment(c)
	return c, nil
}

func (r *CommentRepo) AnswerComment(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) (*models.Comment, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if _, ok := r.store.comments[commentID]; !ok {
		return nil, ErrParentNotFound
	}

	parentID := commentID
	c := &models.Comment{
		ID:              uuid.New(),
		PostID:          postID,
		AuthorID:        authorID,
		ParentCommentID: &parentID,
		Text:            text,
		CreatedAt:       time.Now().UTC(),
	}
	r.store.comments[c.ID] = copyComment(c)
	return c, nil
}

func (r *CommentRepo) GetCommentsPage(ctx context.Context, postID uuid.UUID, parentID *uuid.UUID, limit int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Comment, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	comments := make([]*models.Comment, 0, len(r.store.comments))
	for _, c := range r.store.comments {
		if c.PostID != postID {
			continue
		}
		if parentID == nil {
			if c.ParentCommentID != nil {
				continue
			}
		} else {
			if c.ParentCommentID == nil || *c.ParentCommentID != *parentID {
				continue
			}
		}
		if !commentBefore(c.CreatedAt, c.ID, afterCreatedAt, afterID) {
			continue
		}
		comments = append(comments, copyComment(c))
	}

	sortComments(comments)
	if len(comments) > limit {
		comments = comments[:limit]
	}

	return comments, nil
}

func sortComments(items []*models.Comment) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID.String() > items[j].ID.String()
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
}
