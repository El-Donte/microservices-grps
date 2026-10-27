package main

import (
	"context"
	"errors"
	"fmt"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	lis, err := net.Listen("tcp", ":50051")
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

	if err := serveGRPC(ctx, server, lis); err != nil {
		log.Fatal(err)
	}
}

func serveGRPC(ctx context.Context, grpcServer *grpc.Server, lis net.Listener) error {
	serveDone := make(chan error, 1)
	defer close(serveDone)

	go func() {
		log.Printf("Starting gRPC server on %s", ":50051")
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveDone <- fmt.Errorf("gRPC serve error: %v", err)
		}

		serveDone <- nil
	}()

	select {
	case <-ctx.Done():
		go func() {
			grpcServer.GracefulStop()
		}()

		if err := <-serveDone; err != nil {
			return err
		}

		log.Println("Server exited gracefully")
		return nil

	case err := <-serveDone:
		return err
	}
}
