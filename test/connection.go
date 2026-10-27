package test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/El-Donte/microservices-grps/pkg/api/product-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	productAddr = "localhost:50051"
	cartBase    = "http://localhost:8080"
)

func dialProductService(t *testing.T) product.ProductServiceClient {
	t.Helper()

	conn, err := grpc.NewClient(productAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial product-service: %v", err)
	}

	t.Cleanup(func() { _ = conn.Close() })

	return product.NewProductServiceClient(conn)
}

func dialCartService(t *testing.T, method, path, body string) (int, []byte) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := httpNewRequest(ctx, method, cartBase+path, body)
	if err != nil {
		t.Fatalf("dial cart service: %v", err)
	}

	return httpCall(t, req)
}

func httpNewRequest(ctx context.Context, method, url, body string) (*http.Request, error) {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func httpCall(t *testing.T, req *http.Request) (int, []byte) {
	t.Helper()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("httpCall:", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatal("close response body:", err)
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("read response body:", err)
	}

	return resp.StatusCode, data
}
