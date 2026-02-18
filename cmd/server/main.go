package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/spanner"
	"github.com/product-catalog-service/internal/clock"
	"github.com/product-catalog-service/internal/committer"
	"github.com/product-catalog-service/internal/handler"
	"github.com/product-catalog-service/internal/query"
	"github.com/product-catalog-service/internal/repository"
	"github.com/product-catalog-service/internal/usecase"
	pb "github.com/product-catalog-service/proto/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultPort     = "50051"
	defaultProject  = "test-project"
	defaultInstance = "test-instance"
	defaultDatabase = "test-database"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := getEnv("PORT", defaultPort)
	project := getEnv("SPANNER_PROJECT", defaultProject)
	instance := getEnv("SPANNER_INSTANCE", defaultInstance)
	database := getEnv("SPANNER_DATABASE", defaultDatabase)

	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, database)

	log.Printf("Connecting to Spanner: %s", dbPath)

	spannerClient, err := spanner.NewClient(ctx, dbPath)
	if err != nil {
		log.Fatalf("Failed to create Spanner client: %v", err)
	}
	defer spannerClient.Close()

	productHandler := wireServices(spannerClient)

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, productHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		cancel()
	}()

	log.Printf("Product Catalog Service starting on port %s", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	log.Println("Server stopped")
}

func wireServices(spannerClient *spanner.Client) *handler.Handler {
	clk := clock.NewRealClock()
	comm := committer.NewCommitter(spannerClient)

	productRepo := repository.NewProductRepo(spannerClient)
	outboxRepo := repository.NewOutboxRepo()
	readModel := repository.NewProductReadModel(spannerClient)

	useCases := usecase.NewProductUseCases(productRepo, outboxRepo, comm, clk)
	queries := query.NewProductQueries(readModel, clk)

	return handler.NewHandler(useCases, queries)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
