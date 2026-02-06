package models

import (
	"github.com/google/uuid"
	"time"
)

type Comment struct {
	ID              uuid.UUID
	PostID          uuid.UUID
	AuthorID        uuid.UUID
	ParentCommentID *uuid.UUID
	Text            string
	CreatedAt       time.Time
}
