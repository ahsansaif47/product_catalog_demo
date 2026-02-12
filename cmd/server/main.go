package main

import (
	"context"
	"log"
	"net"
	"os"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc"
	productv1 "product-catalog-service/proto/product/v1"
	"product-catalog-service/internal/services"
	"product-catalog-service/internal/transport/grpc/product"
)

const (
	defaultPort   = "50051"
	defaultSpanner = "projects/test-project/instances/test-instance/databases/product-catalog"
)

func main() {
	port := getEnv("PORT", defaultPort)
	spannerDB := getEnv("SPANNER_DATABASE", defaultSpanner)

	// Initialize Spanner client
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, spannerDB)
	if err != nil {
		log.Fatalf("Failed to create Spanner client: %v", err)
	}
	defer client.Close()

	// Build dependency injection container
	container := services.NewContainer(client)

	// Create gRPC server
	server := grpc.NewServer()

	// Register product service
	productHandler := product.NewHandler(container.ProductHandlers)
	productv1.RegisterProductServiceServer(server, productHandler)

	// Start server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Product Catalog Service starting on port %s", port)
	log.Printf("Spanner database: %s", spannerDB)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
