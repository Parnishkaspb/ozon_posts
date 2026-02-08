package memory

import (
	"context"
	"sort"
	"time"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PostRepo struct {
	store *Store
}

func NewPostRepo(store *Store) *PostRepo {
	return &PostRepo{store: store}
}

func (r *PostRepo) CreatePost(ctx context.Context, authorID uuid.UUID, text string, withoutComment bool) (*models.Post, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	now := time.Now().UTC()
	post := &models.Post{
		ID:             uuid.New(),
		AuthorID:       authorID,
		Text:           text,
		WithoutComment: withoutComment,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	r.store.posts[post.ID] = copyPost(post)
	return post, nil
}

func (r *PostRepo) GetAllPosts(ctx context.Context) ([]*models.Post, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	posts := make([]*models.Post, 0, len(r.store.posts))
	for _, p := range r.store.posts {
		posts = append(posts, copyPost(p))
	}
	sortPosts(posts)
	return posts, nil
}

func (r *PostRepo) GetPostsByID(ctx context.Context, id string) (*models.Post, error) {
	postID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	p, ok := r.store.posts[postID]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return copyPost(p), nil
}

func (r *PostRepo) WithoutComment(ctx context.Context, postID uuid.UUID) (bool, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	p, ok := r.store.posts[postID]
	if !ok {
		return false, pgx.ErrNoRows
	}
	return p.WithoutComment, nil
}

func (r *PostRepo) GetPostsPage(ctx context.Context, first int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Post, bool, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	posts := make([]*models.Post, 0, len(r.store.posts))
	for _, p := range r.store.posts {
		if commentBefore(p.CreatedAt, p.ID, afterCreatedAt, afterID) {
			posts = append(posts, copyPost(p))
		}
	}
	sortPosts(posts)

	hasNext := false
	if len(posts) > first {
		hasNext = true
		posts = posts[:first]
	}

	return posts, hasNext, nil
}

func sortPosts(posts []*models.Post) {
	sort.Slice(posts, func(i, j int) bool {
		if posts[i].CreatedAt.Equal(posts[j].CreatedAt) {
			return posts[i].ID.String() > posts[j].ID.String()
		}
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})
}
