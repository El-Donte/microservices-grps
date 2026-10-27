package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/El-Donte/microservices-grps/iternal/cart/client"
	"github.com/El-Donte/microservices-grps/iternal/cart/repository"
	"github.com/El-Donte/microservices-grps/iternal/cart/service"
	"github.com/El-Donte/microservices-grps/pkg/api/product-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/El-Donte/microservices-grps/docs"
)

// @title cart-service
// @version 1.0
// @host localhost:8080
// @BasePath /
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("dial product-service: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to close connection: %v", err)
		}
	}(conn)

	pc := client.NewGRPCProductClient(product.NewProductServiceClient(conn))
	storage := repository.NewRepository()
	cartService := service.NewCartService(storage, pc)

	if err := serveHTTP(ctx, lis, cartService.Handler()); err != nil {
		log.Fatal(err)
	}
}

func serveHTTP(ctx context.Context, lis net.Listener, handler http.Handler) error {
	httpServer := &http.Server{
		Handler:     handler,
		ReadTimeout: 5 * time.Second,
	}

	serveDone := make(chan error, 1)
	defer close(serveDone)

	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := httpServer.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveDone <- fmt.Errorf("http serve: %v", err)
		}

		serveDone <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			_ = httpServer.Close()
			log.Fatalf("Server shutdown failed: %v", err)
		}

		if err := <-serveDone; err != nil {
			return err
		}

		log.Println("Server exited gracefully")
		return nil

	case err := <-serveDone:
		return err
	}
}
