package posts

import (
	"context"
	"errors"
	"fmt"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) GetAllPosts(ctx context.Context) ([]*models.Post, error) {
	const query = `
		SELECT id, author_id, text, without_comment, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	posts, err := r.collectRows(rows)

	if err != nil {
		return nil, fmt.Errorf("CollectRows: %w", err)
	}

	return posts, nil
}

func (r *Repo) GetPostsByID(ctx context.Context, id string) (*models.Post, error) {
	const query = `
		SELECT id, author_id, text, without_comment, created_at, updated_at
		FROM posts WHERE id = $1
	`

	var post models.Post
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Text,
		&post.WithoutComment,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}

	return &post, nil
}

func (r *Repo) GetPostsByUserID(ctx context.Context, authorID uuid.UUID) ([]*models.Post, error) {
	const query = `
		SELECT id, author_id, text, without_comment, created_at, updated_at
		FROM posts
		WHERE author_id = $1
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	posts, err := r.collectRows(rows)

	if err != nil {
		return nil, fmt.Errorf("CollectRows: %w", err)
	}

	return posts, nil
}

func (r *Repo) GetPostsPage(ctx context.Context, first int, afterCreatedAt *time.Time, afterID *uuid.UUID) ([]*models.Post, bool, error) {
	const query = `
		SELECT id, author_id, text, without_comment, created_at, updated_at
		FROM posts
		WHERE
		  ($1::timestamptz IS NULL AND $2::uuid IS NULL)
		  OR (created_at, id) < ($1::timestamptz, $2::uuid)
		ORDER BY created_at DESC, id DESC
		LIMIT $3;
	`

	limit := first + 1

	rows, err := r.pool.Query(ctx, query, afterCreatedAt, afterID, limit)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	posts := make([]*models.Post, 0, limit)

	for rows.Next() {
		var p models.Post
		if err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Text,
			&p.WithoutComment,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, false, err
		}
		posts = append(posts, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasNext := false
	if len(posts) > first {
		hasNext = true
		posts = posts[:first]
	}

	return posts, hasNext, nil
}

func (r *Repo) collectRows(rows pgx.Rows) ([]*models.Post, error) {
	posts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.Post, error) {
		p := new(models.Post)
		return p, row.Scan(&p.ID, &p.AuthorID, &p.Text, &p.WithoutComment, &p.CreatedAt, &p.UpdatedAt)
	})

	return posts, err
}

func (r *Repo) CreatePost(ctx context.Context, authorID uuid.UUID, text string, withoutComment bool) (*models.Post, error) {
	const query = `
		INSERT INTO posts (author_id, text, without_comment)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	var post models.Post
	err := r.pool.QueryRow(ctx, query, authorID, text, withoutComment).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("QueryRow Scan: %w", err)
	}
	post.AuthorID = authorID
	post.Text = text
	post.WithoutComment = withoutComment

	return &post, nil
}

func (r *Repo) WithoutComment(ctx context.Context, postID uuid.UUID) (bool, error) {
	const query = `SELECT without_comment FROM posts WHERE id = $1`
	var withoutComment bool
	err := r.pool.QueryRow(ctx, query, postID).Scan(&withoutComment)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, pgx.ErrNoRows
	}

	return withoutComment, err
}
