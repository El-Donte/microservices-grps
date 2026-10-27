package client

import (
	"context"
	"sort"
	"sync"

	"github.com/El-Donte/microservices-grps/iternal/cart/domain"
	"github.com/El-Donte/microservices-grps/pkg/api/product-service"
)

type ProductClient interface {
	GetProduct(ctx context.Context, sku uint64) (name string, price uint32, err error)
	FetchProducts(ctx context.Context, skus []uint64) ([]domain.Item, error)
}

type GRPCClient struct {
	client product.ProductServiceClient
}

func NewGRPCProductClient(client product.ProductServiceClient) *GRPCClient {
	return &GRPCClient{client: client}
}

func (c *GRPCClient) GetProduct(ctx context.Context, sku uint64) (string, uint32, error) {
	response, err := c.client.GetProduct(ctx, &product.GetProductRequest{Sku: int64(sku)})
	if err != nil {
		return "", 0, err
	}
	if response.GetPrice() < 0 {
		return response.GetName(), 0, nil
	}
	return response.GetName(), uint32(response.GetPrice()), nil
}

func (c *GRPCClient) FetchProducts(ctx context.Context, skus []uint64) ([]domain.Item, error) {
	if len(skus) == 0 {
		return nil, nil
	}

	type result struct {
		name  string
		price uint32
		err   error
	}

	results := make([]result, len(skus))

	var wg sync.WaitGroup
	wg.Add(len(skus))

	for i, sku := range skus {
		go func() {
			defer wg.Done()
			name, price, err := c.GetProduct(ctx, sku)
			results[i] = result{name: name, price: price, err: err}
		}()
	}
	wg.Wait()

	resp := make([]domain.Item, 0, len(skus))
	for i, sku := range skus {
		if results[i].err != nil {
			return nil, results[i].err
		}
		resp = append(resp, domain.Item{SKU: sku, Name: results[i].name, Price: results[i].price})
	}

	sort.Slice(resp, func(i, j int) bool { return resp[i].SKU < resp[j].SKU })
	return resp, nil
}
