package helpergraph

import (
	"context"
	"encoding/base64"
	"fmt"
	graphdataloader "github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/dataloader"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/model"
	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
	"github.com/graph-gophers/dataloader"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func ResolveAuthor(ctx context.Context, authorID string) (*model.User, error) {
	if authorID == "" {
		return nil, fmt.Errorf("authorId is empty")
	}

	lds, ok := graphdataloader.FromContext(ctx)
	if !ok || lds.UsersByIDs == nil {
		return nil, graphdataloader.ErrNotInjected
	}

	thunk := lds.UsersByIDs.Load(ctx, dataloader.StringKey(authorID))
	data, err := thunk()
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("author not found")
	}

	u := data.(*servicepb.User)
	return &model.User{
		ID:      u.GetId(),
		Name:    u.GetName(),
		Surname: u.GetSurname(),
	}, nil
}

func MakeCursor(ts *timestamppb.Timestamp, id string) string {
	raw := ts.AsTime().UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}
