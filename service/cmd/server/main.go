package main

import (
	"context"
	"github.com/Parnishkaspb/ozon_posts/internal/auth"
	grpchandlers "github.com/Parnishkaspb/ozon_posts/internal/transport/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Parnishkaspb/ozon_posts/internal/app"
	"github.com/Parnishkaspb/ozon_posts/internal/config"
	servicepb "github.com/Parnishkaspb/ozon_posts_proto/gen/service/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.MustLoad("config/config.yaml")

	port := strconv.Itoa(cfg.GRPC.Port)
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("grpc listen :%s error: %v", port, err)
	}

	jwtService := auth.New(cfg.JWT.Secret, cfg.JWT.TTL)

	a, err := app.New(context.Background(), cfg.PostgresDSN(), jwtService)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	h := grpchandlers.New(a)
	grpcServer := grpc.NewServer()
	servicepb.RegisterAuthServiceServer(grpcServer, h)
	servicepb.RegisterUserServiceServer(grpcServer, h)
	servicepb.RegisterPostServiceServer(grpcServer, h)
	servicepb.RegisterCommentServiceServer(grpcServer, h)

	reflection.Register(grpcServer)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("gRPC server listening on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown: stopping gRPC server...")

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("shutdown: gRPC stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Println("shutdown: timeout -> force stop")
		grpcServer.Stop()
	}
	_ = servicepb.File_service_v1_service_proto
}
