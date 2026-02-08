package users

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
)

type mockUsersRepo struct {
	allUsers    []*models.User
	usersByID   *models.User
	usersByIDs  []*models.User
	err         error
	calledAll   int
	calledByID  int
	calledByIDs int
	gotIDs      []string
	gotUserID   uuid.UUID
}

func (m *mockUsersRepo) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	m.calledAll++
	return m.allUsers, m.err
}

func (m *mockUsersRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	m.calledByID++
	m.gotUserID = userID
	return m.usersByID, m.err
}

func (m *mockUsersRepo) GetUsersByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	m.calledByIDs++
	m.gotIDs = append([]string(nil), ids...)
	return m.usersByIDs, m.err
}

func TestUserService_GetUsersByIds(t *testing.T) {
	ctx := context.Background()
	expectedAll := []*models.User{{ID: uuid.New()}}
	expectedByIDs := []*models.User{{ID: uuid.New()}}
	repoErr := errors.New("repo err")

	tests := []struct {
		name           string
		ids            []string
		repo           *mockUsersRepo
		want           []*models.User
		wantErr        error
		wantAllCalls   int
		wantByIDsCalls int
		wantPassedIDs  []string
	}{
		{
			name:         "empty ids uses get all",
			ids:          nil,
			repo:         &mockUsersRepo{allUsers: expectedAll},
			want:         expectedAll,
			wantAllCalls: 1,
		},
		{
			name:           "non-empty ids uses get by ids",
			ids:            []string{"1", "2"},
			repo:           &mockUsersRepo{usersByIDs: expectedByIDs},
			want:           expectedByIDs,
			wantByIDsCalls: 1,
			wantPassedIDs:  []string{"1", "2"},
		},
		{
			name:           "repo error propagated",
			ids:            []string{"1"},
			repo:           &mockUsersRepo{err: repoErr},
			wantErr:        repoErr,
			wantByIDsCalls: 1,
			wantPassedIDs:  []string{"1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewUserService(tt.repo)
			got, err := svc.GetUsersByIds(ctx, tt.ids)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected users")
			}
			if tt.repo.calledAll != tt.wantAllCalls || tt.repo.calledByIDs != tt.wantByIDsCalls {
				t.Fatalf("unexpected repo calls")
			}
			if !reflect.DeepEqual(tt.repo.gotIDs, tt.wantPassedIDs) {
				t.Fatalf("unexpected ids: got %v want %v", tt.repo.gotIDs, tt.wantPassedIDs)
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	u := &models.User{ID: uuid.New()}
	repoErr := errors.New("repo err")

	tests := []struct {
		name    string
		repo    *mockUsersRepo
		userID  uuid.UUID
		want    *models.User
		wantErr error
	}{
		{name: "success", repo: &mockUsersRepo{usersByID: u}, userID: u.ID, want: u},
		{name: "repo error", repo: &mockUsersRepo{err: repoErr}, userID: uuid.New(), wantErr: repoErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewUserService(tt.repo)
			got, err := svc.GetUserByID(ctx, tt.userID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected user")
			}
			if tt.repo.calledByID != 1 || tt.repo.gotUserID != tt.userID {
				t.Fatalf("unexpected repo call")
			}
		})
	}
}
