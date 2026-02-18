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
	"github.com/product-catalog-service/internal/app"
	pb "github.com/product-catalog-service/proto/product/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultPort      = "50051"
	defaultProject   = "test-project"
	defaultInstance  = "test-instance"
	defaultDatabase  = "test-database"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get configuration from environment
	port := getEnv("PORT", defaultPort)
	project := getEnv("SPANNER_PROJECT", defaultProject)
	instance := getEnv("SPANNER_INSTANCE", defaultInstance)
	database := getEnv("SPANNER_DATABASE", defaultDatabase)

	// Build Spanner database path
	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, database)

	log.Printf("Connecting to Spanner: %s", dbPath)

	// Create Spanner client
	spannerClient, err := spanner.NewClient(ctx, dbPath)
	if err != nil {
		log.Fatalf("Failed to create Spanner client: %v", err)
	}
	defer spannerClient.Close()

	// Initialize services
	svc := app.NewServices(spannerClient)
	defer svc.Close()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register ProductService
	pb.RegisterProductServiceServer(grpcServer, svc.ProductHandler)

	// Enable reflection for debugging with tools like grpcurl
	reflection.Register(grpcServer)

	// Start listening
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	// Handle graceful shutdown
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

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
