package graph

import servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Auth servicepb.AuthServiceClient
	User servicepb.UserServiceClient
	Post servicepb.PostServiceClient
}
