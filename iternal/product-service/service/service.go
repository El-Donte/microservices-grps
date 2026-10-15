package service

import (
	"context"

	"github.com/El-Donte/microservices-grps/iternal/product-service/domain"
	"github.com/El-Donte/microservices-grps/iternal/product-service/repository"
	pbProduct "github.com/El-Donte/microservices-grps/pkg/api/product-service"
)

type ProductService struct {
	pbProduct.UnimplementedProductServiceServer
	repo *repository.ProductRepository
}

func NewProductService(path string) (*ProductService, error) {
	repo, err := repository.NewProductRepository(path)
	if err != nil {
		return nil, err
	}

	return &ProductService{repo: repo}, nil
}

func (p ProductService) GetProduct(ctx context.Context, request *pbProduct.GetProductRequest) (*pbProduct.GetProductResponse, error) {
	sku := request.GetSku()
	if sku <= 0 {
		return nil, domain.SkuMustBePositive
	}

	product, err := p.repo.GetProduct(uint64(sku))
	if err != nil {
		return nil, domain.ProductNotFoundError
	}

	return &pbProduct.GetProductResponse{
		Sku:   int64(product.SKU),
		Name:  product.Name,
		Price: int32(product.Price),
	}, nil
}

func (p ProductService) ListProducts(ctx context.Context, request *pbProduct.ListProductsRequest) (*pbProduct.ListProductsResponse, error) {
	page := request.GetPage()
	if page <= 0 {
		page = 1
	}
	perPage := request.GetPerPage()
	if perPage <= 0 {
		perPage = 10
	}

	list := p.repo.ListProducts()
	total := uint32(len(list))

	start := (page - 1) * perPage
	end := min(start+perPage, total)

	if start >= total {
		return &pbProduct.ListProductsResponse{
			Total:   total,
			Page:    page,
			PerPage: perPage,
		}, nil
	}

	pageProducts := list[start:end]

	products := make([]*pbProduct.Product, len(pageProducts))
	for i, product := range pageProducts {
		products[i] = &pbProduct.Product{
			Sku:   int64(product.SKU),
			Name:  product.Name,
			Price: int32(product.Price),
		}
	}

	return &pbProduct.ListProductsResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	}, nil
}
