package unit_test

import (
	"context"
	"testing"

	"github.com/El-Donte/microservices-grps/iternal/product-service/service"
	"github.com/El-Donte/microservices-grps/pkg/api/product-service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newProductService(t *testing.T) *service.ProductService {
	t.Helper()

	productService, err := service.NewProductService("../../../products.json")

	if err != nil {
		t.Fatalf("could not create product service: %v", err)
	}

	return productService
}

func TestGetProduct(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		sku     int64
		wantErr codes.Code
		want    *product.GetProductResponse
	}{
		{"happy path", 1076963, codes.OK,
			&product.GetProductResponse{Sku: 1076963, Name: "Теория нравственных чувств | Смит Адам", Price: 3379}},
		{"zero sku", 0, codes.InvalidArgument, nil},
		{"negative sku", -7, codes.InvalidArgument, nil},
		{"not found", 1, codes.NotFound, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svc := newProductService(t)

			resp, err := svc.GetProduct(context.Background(),
				&product.GetProductRequest{Sku: testCase.sku})

			if testCase.wantErr != codes.OK {
				require.Error(t, err)
				require.Equal(t, testCase.wantErr, status.Code(err))
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.want, resp)
		})
	}
}

func TestListProducts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		page, perPage     uint32
		wantLen           int
		wantPage, wantPer uint32
	}{
		{"defaults", 0, 0, 5, 1, 10},
		{"page 1 size 2", 1, 2, 2, 1, 2},
		{"page 2 size 2", 2, 2, 2, 2, 2},
		{"past the end", 10, 2, 0, 10, 2},
		{"per_page capped", 1, 500, 5, 1, 10},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svc := newProductService(t)

			resp, err := svc.ListProducts(context.Background(),
				&product.ListProductsRequest{Page: testCase.page, PerPage: testCase.perPage})

			require.NoError(t, err)
			require.Len(t, resp.GetProducts(), testCase.wantLen)
			require.Equal(t, testCase.wantPage, resp.GetPage())
			require.Equal(t, testCase.wantPer, resp.GetPerPage())
		})
	}
}

func TestListProductsOrdered(t *testing.T) {
	t.Parallel()
	productService := newProductService(t)

	resp, err := productService.ListProducts(context.Background(),
		&product.ListProductsRequest{Page: 1, PerPage: 10})

	require.NoError(t, err)
	require.Equal(t, []int64{2008, 1076963, 1148162}, []int64{
		resp.GetProducts()[0].GetSku(),
		resp.GetProducts()[1].GetSku(),
		resp.GetProducts()[2].GetSku(),
	})
}
