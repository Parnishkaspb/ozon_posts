package subscriptions

import (
	"sync"

	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/model"
)

type Subscription struct {
	mu   sync.RWMutex
	subs map[string]map[chan *model.Comment]struct{}
}

func New() *Subscription {
	return &Subscription{
		subs: make(map[string]map[chan *model.Comment]struct{}),
	}
}

func (ps *Subscription) Subscribe(postID string) chan *model.Comment {
	ch := make(chan *model.Comment, 16)

	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.subs[postID] == nil {
		ps.subs[postID] = make(map[chan *model.Comment]struct{})
	}
	ps.subs[postID][ch] = struct{}{}

	return ch
}

func (ps *Subscription) Unsubscribe(postID string, ch chan *model.Comment) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if m := ps.subs[postID]; m != nil {
		delete(m, ch)
		if len(m) == 0 {
			delete(ps.subs, postID)
		}
	}
	close(ch)
}

func (ps *Subscription) Publish(postID string, c *model.Comment) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for ch := range ps.subs[postID] {
		select {
		case ch <- c:
		default:
		}
	}
}
