package grpc

import (
	"errors"
	"testing"

	"github.com/Parnishkaspb/ozon_posts/internal/services/comments"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGrpcErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{
			name: "post required",
			err:  comments.ErrPostIDRequired,
			code: codes.InvalidArgument,
		},
		{
			name: "invalid cursor",
			err:  comments.ErrInvalidCursor,
			code: codes.InvalidArgument,
		},
		{
			name: "internal",
			err:  errors.New("internal error"),
			code: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, ok := status.FromError(grpcErr(tt.err))
			if !ok {
				t.Fatalf("expected grpc status")
			}
			if st.Code() != tt.code {
				t.Fatalf("expected %v, got %v", tt.code, st.Code())
			}
		})
	}
}
