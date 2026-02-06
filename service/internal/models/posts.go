package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID             uuid.UUID `json:"id"`
	AuthorID       uuid.UUID `json:"author_id"`
	Text           string    `json:"text"`
	WithoutComment bool      `json:"without_comment"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
