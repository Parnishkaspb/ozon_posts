package graph

import (
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/subscriptions"
	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	AuthSvc    servicepb.AuthServiceClient
	UserSvc    servicepb.UserServiceClient
	PostSvc    servicepb.PostServiceClient
	CommentSvc servicepb.CommentServiceClient

	SubSvc *subscriptions.Subscription
}
