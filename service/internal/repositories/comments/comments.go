package comments

import (
	"context"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
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

func (r *Repo) GetCommentsPage(ctx context.Context, postID uuid.UUID, parentID *uuid.UUID, limit int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Comment, error) {
	const query = `
		SELECT id, post_id, author_id, parent_id, text, created_at
		FROM comments
		WHERE
			post_id = $1
			AND (
				($2::uuid IS NULL AND parent_id IS NULL)
				OR
				($2::uuid IS NOT NULL AND parent_id = $2::uuid)
			)
			AND (
				($3::timestamptz IS NULL AND $4::uuid IS NULL)
				OR
				(created_at, id) < ($3::timestamptz, $4::uuid)
			)
		ORDER BY created_at DESC, id DESC
		LIMIT $5;
	`

	rows, err := r.pool.Query(ctx, query, postID, parentID, afterCreatedAt, afterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*models.Comment, 0, limit)

	for rows.Next() {
		var c models.Comment
		var pID *uuid.UUID

		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.AuthorID,
			&pID,
			&c.Text,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}

		c.ParentCommentID = pID
		comments = append(comments, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
