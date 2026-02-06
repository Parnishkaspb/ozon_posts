package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/auth"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/config"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph"
	"github.com/Parnishkaspb/ozon_posts_graphql/internal/graph/generated"
	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"os"
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

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{
			Auth: authClient,
			User: userClient,
			Post: postClient,
		},
	}))

	jwtService := auth.New(cfg.JWT.Secret, cfg.JWT.TTL)

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", auth.AuthMiddleware(jwtService, srv))

	log.Printf("GraphQL started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
