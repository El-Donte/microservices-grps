package main

import (
	"context"
	"errors"
	"log"
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
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("dial product-service: %v", err)
	}
	defer conn.Close()

	pc := client.NewGRPCProductClient(product.NewProductServiceClient(conn))
	storage := repository.NewRepository()
	srv := service.NewService(storage, pc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv.Handler(),
	}

	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, stopping server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server exited gracefully")
}
