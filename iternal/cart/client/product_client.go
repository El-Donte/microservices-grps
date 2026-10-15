package client

import (
	"context"

	"github.com/El-Donte/microservices-grps/pkg/api/product-service"
)

type GRPsClient struct {
	client product.ProductServiceClient
}

func NewGRPCProductClient(client product.ProductServiceClient) *GRPsClient {
	return &GRPsClient{client: client}
}

func (c *GRPsClient) GetProduct(ctx context.Context, sku uint64) (string, uint32, error) {
	response, err := c.client.GetProduct(ctx, &product.GetProductRequest{Sku: int64(sku)})
	if err != nil {
		return "", 0, err
	}
	if response.GetPrice() < 0 {
		return response.GetName(), 0, nil
	}
	return response.GetName(), uint32(response.GetPrice()), nil
}
