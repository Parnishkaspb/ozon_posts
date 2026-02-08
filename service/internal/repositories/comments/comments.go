package comments

import (
	"context"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) CreateComment(ctx context.Context, text string, authorID, postID uuid.UUID) (*models.Comment, error) {
	const query = `
		INSERT INTO comments (post_id, author_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, post_id, author_id, parent_id, text, created_at;
	`

	return r.exec(ctx, query, postID, authorID, text)
}

func (r *Repo) AnswerComment(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) (*models.Comment, error) {
	const query = `
		INSERT INTO comments (post_id, author_id, parent_id, text)
		VALUES ($1, $2, $3, $4)
		RETURNING id, post_id, author_id, parent_id, text, created_at;
	`

	return r.exec(ctx, query, postID, authorID, commentID, text)
}

func (r *Repo) exec(ctx context.Context, query string, args ...any) (*models.Comment, error) {
	var c models.Comment
	var parentID *uuid.UUID

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&c.ID,
		&c.PostID,
		&c.AuthorID,
		&parentID,
		&c.Text,
		&c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	c.ParentCommentID = parentID
	return &c, nil
}
