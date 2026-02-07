package dataloader

import (
	"context"
	"errors"
	"time"

	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
	"github.com/graph-gophers/dataloader"
)

var ErrNotInjected = errors.New("dataloader not injected")

type key struct{}

type Loaders struct {
	UsersByIDs *dataloader.Loader
}

func New(userSvc servicepb.UserServiceClient) *Loaders {
	return &Loaders{
		UsersByIDs: dataloader.NewBatchedLoader(
			batchUsers(userSvc),
			dataloader.WithWait(2*time.Millisecond),
			dataloader.WithBatchCapacity(200),
		),
	}
}

func Inject(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, key{}, loaders)
}

func FromContext(ctx context.Context) (*Loaders, bool) {
	lds, ok := ctx.Value(key{}).(*Loaders)
	return lds, ok
}

func batchUsers(userSvc servicepb.UserServiceClient) dataloader.BatchFunc {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		uniq := make([]string, 0, len(keys))
		seen := make(map[string]struct{}, len(keys))
		for _, k := range keys {
			id := k.String()
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			uniq = append(uniq, id)
		}

		if len(uniq) == 0 {
			out := make([]*dataloader.Result, len(keys))
			for i := range out {
				out[i] = &dataloader.Result{Data: nil, Error: nil}
			}
			return out
		}

		rpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		resp, err := userSvc.GetUsers(rpcCtx, &servicepb.GetUsersRequest{Ids: uniq})
		if err != nil {
			out := make([]*dataloader.Result, len(keys))
			for i := range out {
				out[i] = &dataloader.Result{Error: err}
			}
			return out
		}

		m := make(map[string]*servicepb.User, len(resp.GetUsers()))
		for _, u := range resp.GetUsers() {
			if u == nil || u.GetId() == "" {
				continue
			}
			m[u.GetId()] = u
		}

		out := make([]*dataloader.Result, len(keys))
		for i, k := range keys {
			id := k.String()
			if id == "" {
				out[i] = &dataloader.Result{Data: nil, Error: nil}
				continue
			}
			out[i] = &dataloader.Result{Data: m[id], Error: nil} // nil = not found
		}
		return out
	}
}
