package unit

import (
	"context"
	"sort"

	"github.com/El-Donte/microservices-grps/iternal/cart/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FakeProductService struct {
	products map[uint64]struct {
		Name  string
		Price uint32
	}
}

func (f FakeProductService) FetchProducts(ctx context.Context, skus []uint64) ([]domain.Item, error) {
	list := make([]domain.Item, 0, len(skus))
	for _, sku := range skus {
		product, ok := f.products[sku]
		if ok {
			list = append(list, domain.Item{SKU: sku, Name: product.Name, Price: product.Price})
		}
	}

	sort.Slice(list, func(i, j int) bool { return list[i].SKU < list[j].SKU })
	return list, nil
}

func NewFakeProductService() *FakeProductService {
	return &FakeProductService{
		products: map[uint64]struct {
			Name  string
			Price uint32
		}{
			1: {"book", 100},
			2: {"lamp", 200},
			3: {"pen", 300},
		},
	}
}

func (f FakeProductService) GetProduct(_ context.Context, sku uint64) (name string, price uint32, err error) {
	response, ok := f.products[sku]
	if !ok {
		return "", 0, status.Errorf(codes.NotFound, "not found")
	}

	return response.Name, response.Price, nil
}
