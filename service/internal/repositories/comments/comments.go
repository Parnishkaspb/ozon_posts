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
		INSERT INTO comments (text, author_id, post_id) VALUES ($1, $2, $3) RETURNING id, created_at;
	`
	var comment models.Comment
	err := r.pool.QueryRow(ctx, query, text, authorID, postID).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		return &models.Comment{}, err
	}
	comment.Text = text
	comment.AuthorID = authorID
	comment.PostID = postID

	return &comment, nil
}

func (r *Repo) AnswerComment(ctx context.Context, text string, authorID, postID, commentID uuid.UUID) error {
	const query = `
		INSERT INTO comments (text, author_id, parent_comment_id, post_id) VALUES ($1, $2, $3, $4)
	`

	return r.exec(ctx, query, text, authorID, postID, commentID)
}

func (r *Repo) exec(ctx context.Context, query string, args ...any) error {
	_, err := r.pool.Exec(ctx, query, args)

	if err != nil {
		return err
	}

	return nil
}
