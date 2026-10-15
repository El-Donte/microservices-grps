package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/El-Donte/microservices-grps/iternal/product-service/service"
	pbProduct "github.com/El-Donte/microservices-grps/pkg/api/product-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)

	productService, err := service.NewProductService("products.json")
	if err != nil {
		log.Fatalf("Failed to create product service: %v", err)
	}
	pbProduct.RegisterProductServiceServer(server, productService)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Starting gRPC server on %s", ":50051")
		if err := server.Serve(listen); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, stopping server...")

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	log.Println("Server exited gracefully")
}
