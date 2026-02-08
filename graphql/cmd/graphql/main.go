//package main
//
//import (
//	"github.com/99designs/gqlgen/graphql/handler"
//	"github.com/Parnishkaspb/ozon_posts_graphql/internal/auth"
//	"github.com/Parnishkaspb/ozon_posts_graphql/internal/config"
//	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph"
//	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/dataloader"
//	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/generated"
//	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/subscriptions"
//	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/credentials/insecure"
//	"log"
//	"net/http"
//	"os"
//)
//
//func main() {
//	cfgPath := os.Getenv("CONFIG_PATH")
//	if cfgPath == "" {
//		cfgPath = "config/config.yaml"
//	}
//	cfg := config.MustLoad(cfgPath)
//	port := "8080"
//	grpcTarget := "service:9090"
//
//	conn, err := grpc.Dial(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		log.Fatalf("grpc dial %s: %v", grpcTarget, err)
//	}
//	defer conn.Close()
//
//	authClient := servicepb.NewAuthServiceClient(conn)
//	userClient := servicepb.NewUserServiceClient(conn)
//	postClient := servicepb.NewPostServiceClient(conn)
//	commentClient := servicepb.NewCommentServiceClient(conn)
//	subService := subscriptions.New()
//
//	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
//		Resolvers: &graph.Resolver{
//			AuthSvc:    authClient,
//			UserSvc:    userClient,
//			PostSvc:    postClient,
//			CommentSvc: commentClient,
//			SubSvc:     subService,
//		},
//	}))
//
//	jwtService := auth.New(cfg.JWT.Secret, cfg.JWT.TTL)
//
//	mux := http.NewServeMux()
//	queryHandler := auth.AuthMiddleware(jwtService, srv)
//
//	mux.Handle("/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		lds := dataloader.New(userClient)
//		ctx := dataloader.Inject(r.Context(), lds)
//		queryHandler.ServeHTTP(w, r.WithContext(ctx))
//	}))
//
//	log.Printf("GraphQL started on :%s", port)
//	log.Fatal(http.ListenAndServe(":"+port, mux))
//}

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"

	"github.com/Parnishkaspb/ozon_posts_graphql/internal/auth"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/config"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/dataloader"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/generated"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/subscriptions"

	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/config.yaml"
	}
	cfg := config.MustLoad(cfgPath)

	port := "8080"
	grpcTarget := "service:9090"

	conn, err := grpc.Dial(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc dial %s: %v", grpcTarget, err)
	}
	defer conn.Close()

	authClient := servicepb.NewAuthServiceClient(conn)
	userClient := servicepb.NewUserServiceClient(conn)
	postClient := servicepb.NewPostServiceClient(conn)
	commentClient := servicepb.NewCommentServiceClient(conn)

	subService := subscriptions.New()
	jwtService := auth.New(cfg.JWT.Secret, cfg.JWT.TTL)

	// gqlgen server
	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{
			AuthSvc:    authClient,
			UserSvc:    userClient,
			PostSvc:    postClient,
			CommentSvc: commentClient,
			SubSvc:     subService,
		},
	}))

	// ВАЖНО: транспорты для subscription + query
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		KeepAlivePingInterval: 15 * time.Second,
	})

	mux := http.NewServeMux()

	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))

	mux.Handle("/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1) dataloader в контекст
		lds := dataloader.New(userClient)
		ctx := dataloader.Inject(r.Context(), lds)

		// 2) auth middleware (важно: он должен корректно пропускать WS)
		// Если твой middleware ломает WS — см. ниже примечание
		next := auth.AuthMiddleware(jwtService, srv)

		next.ServeHTTP(w, r.WithContext(ctx))
	}))

	log.Printf("GraphQL started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
