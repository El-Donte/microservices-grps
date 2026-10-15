package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/El-Donte/microservices-grps/iternal/product-service/domain"
)

type ProductRepository struct {
	mu       sync.RWMutex
	database map[uint64]domain.Product
}

func NewProductRepository(path string) (*ProductRepository, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read product repository: %w", err)
	}

	var decoded map[string]struct {
		Name  string `json:"name"`
		Price int32  `json:"price"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product repository: %w", err)
	}

	database := make(map[uint64]domain.Product, len(decoded))

	for skuStr, product := range decoded {
		sku, err := strconv.ParseUint(skuStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sku: %w", err)
		}
		if product.Price <= 0 {
			return nil, fmt.Errorf("invalid product price for sku: %d", sku)
		}
		if product.Name == "" {
			return nil, fmt.Errorf("invalid product name for sku: %s", sku)
		}

		database[sku] = domain.Product{
			SKU:   sku,
			Name:  product.Name,
			Price: uint32(product.Price),
		}
	}

	return &ProductRepository{database: database}, nil
}

func (r *ProductRepository) GetProduct(sku uint64) (domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, ok := r.database[sku]
	if !ok {
		return domain.Product{}, fmt.Errorf("product not found")
	}
	return product, nil
}

func (r *ProductRepository) ListProducts() []domain.Product {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]domain.Product, 0, len(r.database))
	for _, product := range r.database {
		list = append(list, product)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].SKU < list[j].SKU
	})

	return list
}
