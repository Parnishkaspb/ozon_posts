package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Parnishkaspb/ozon_posts/internal/auth"
	"github.com/Parnishkaspb/ozon_posts/internal/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.Config
		wantErr     error
		wantPoolNil bool
	}{
		{
			name: "memory driver",
			cfg: &config.Config{
				Storage: config.StorageConfig{Driver: "memory"},
				JWT:     config.Token{Secret: "secret", TTL: time.Minute},
			},
			wantPoolNil: true,
		},
		{
			name:    "unknown driver",
			cfg:     &config.Config{Storage: config.StorageConfig{Driver: "unknown"}},
			wantErr: ErrUnknownStorageDriver,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := New(context.Background(), tt.cfg, auth.New("secret", time.Minute))
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a == nil || a.PostSRV == nil || a.CommentSRV == nil || a.UserSRV == nil || a.Auth == nil {
				t.Fatalf("app not initialized")
			}
			if (a.Pool == nil) != tt.wantPoolNil {
				t.Fatalf("unexpected pool state")
			}
		})
	}
}
