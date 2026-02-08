package memory

import (
	"sync"
	"time"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
)

type Store struct {
	mu       sync.RWMutex
	users    map[uuid.UUID]*models.User
	posts    map[uuid.UUID]*models.Post
	comments map[uuid.UUID]*models.Comment
}

func NewStore() *Store {
	s := &Store{
		users:    make(map[uuid.UUID]*models.User),
		posts:    make(map[uuid.UUID]*models.Post),
		comments: make(map[uuid.UUID]*models.Comment),
	}

	seedID := uuid.New()
	s.users[seedID] = &models.User{
		ID:       seedID,
		Login:    "Ivan",
		Password: "MoscowNeverSleep",
		Name:     "Иван",
		Surname:  "Грозный",
	}

	return s
}

func copyUser(u *models.User) *models.User {
	if u == nil {
		return nil
	}
	cp := *u
	return &cp
}

func copyPost(p *models.Post) *models.Post {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

func copyComment(c *models.Comment) *models.Comment {
	if c == nil {
		return nil
	}
	cp := *c
	if c.ParentCommentID != nil {
		pid := *c.ParentCommentID
		cp.ParentCommentID = &pid
	}
	return &cp
}

func commentBefore(aTime time.Time, aID uuid.UUID, bTime *time.Time, bID *uuid.UUID) bool {
	if bTime == nil || bID == nil {
		return true
	}
	if aTime.Before(*bTime) {
		return true
	}
	if aTime.Equal(*bTime) {
		return aID.String() < bID.String()
	}
	return false
}
