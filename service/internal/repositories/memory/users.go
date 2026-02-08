package memory

import (
	"context"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	pgusers "github.com/Parnishkaspb/ozon_posts/internal/repositories/users"
	"github.com/google/uuid"
)

type UserRepo struct {
	store *Store
}

func NewUserRepo(store *Store) *UserRepo {
	return &UserRepo{store: store}
}

func (r *UserRepo) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	out := make([]*models.User, 0, len(r.store.users))
	for _, u := range r.store.users {
		out = append(out, copyUser(u))
	}
	return out, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	u, ok := r.store.users[userID]
	if !ok {
		return nil, nil
	}
	return copyUser(u), nil
}

func (r *UserRepo) GetUsersByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	out := make([]*models.User, 0, len(ids))
	for _, raw := range ids {
		id, err := uuid.Parse(raw)
		if err != nil {
			continue
		}
		u, ok := r.store.users[id]
		if !ok {
			continue
		}
		out = append(out, copyUser(u))
	}
	return out, nil
}

func (r *UserRepo) GetUserByLoginPassword(ctx context.Context, login, password string) (*models.User, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	for _, u := range r.store.users {
		if u.Login == login && u.Password == password {
			return copyUser(u), nil
		}
	}

	return nil, pgusers.ErrUserNotFound
}
