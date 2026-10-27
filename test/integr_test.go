package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/El-Donte/microservices-grps/pkg/api/product-service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProductGetProduct(t *testing.T) {
	productClient := dialProductService(t)
	ctx, c := context.WithTimeout(context.Background(), 5*time.Second)
	defer c()

	t.Run("happy path", func(t *testing.T) {
		resp, err := productClient.GetProduct(ctx, &product.GetProductRequest{Sku: 1076963})
		require.NoError(t, err)
		require.Equal(t, int64(1076963), resp.GetSku())
		require.NotEmpty(t, resp.GetName())
		require.Greater(t, resp.GetPrice(), int32(0))
	})

	t.Run("zero sku", func(t *testing.T) {
		_, err := productClient.GetProduct(ctx, &product.GetProductRequest{Sku: 0})
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("negative sku", func(t *testing.T) {
		_, err := productClient.GetProduct(ctx, &product.GetProductRequest{Sku: -10})
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("not found", func(t *testing.T) {
		_, err := productClient.GetProduct(ctx, &product.GetProductRequest{Sku: 999999999})
		require.Error(t, err)
		require.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestCartToProductCalls(t *testing.T) {
	userId := time.Now().Second()

	t.Run("empty cart", func(t *testing.T) {
		code, _ := dialCartService(t, "GET", fmt.Sprintf("/user/%d/cart", userId), "")
		require.Equal(t, http.StatusNotFound, code)
	})

	t.Run("add existing product", func(t *testing.T) {
		code, body := dialCartService(t, "POST", fmt.Sprintf("/user/%d/cart/1076963", userId),
			`{"count":3}`)
		require.Equal(t, http.StatusOK, code, string(body))
	})

	t.Run("add not existing sku", func(t *testing.T) {
		code, body := dialCartService(t, "POST", fmt.Sprintf("/user/%d/cart/10", userId),
			`{"count":3}`)
		require.Equal(t, http.StatusPreconditionFailed, code, string(body))
	})

	t.Run("invalid body", func(t *testing.T) {
		code, body := dialCartService(t, "POST", fmt.Sprintf("/user/%d/cart/1076963", userId),
			`{"count": 0}`)
		require.Equal(t, http.StatusBadRequest, code, string(body))
	})

	t.Run("invalid post userId", func(t *testing.T) {
		code, body := dialCartService(t, "POST", fmt.Sprintf("/user/%d/cart/1076963", 0),
			`{"count": 1}`)
		require.Equal(t, http.StatusBadRequest, code, string(body))
	})

	t.Run("invalid get userId", func(t *testing.T) {
		code, body := dialCartService(t, "GET", fmt.Sprintf("/user/%d/cart", 0), "")
		require.Equal(t, http.StatusBadRequest, code, string(body))
	})

	t.Run("get cart contents", func(t *testing.T) {
		code, body := dialCartService(t, "GET", fmt.Sprintf("/user/%d/cart", userId), "")
		require.Equal(t, http.StatusOK, code, string(body))
		var resp struct {
			Items []struct {
				SKU   int64  `json:"sku"`
				Name  string `json:"name"`
				Count uint32 `json:"count"`
				Price uint32 `json:"price"`
			} `json:"items"`
			TotalPrice uint32 `json:"total_price"`
		}
		require.NoError(t, json.Unmarshal(body, &resp))
		require.Len(t, resp.Items, 1)
		require.Equal(t, int64(1076963), resp.Items[0].SKU)
		require.Equal(t, uint32(3), resp.Items[0].Count)
		require.Equal(t, uint32(10137), resp.TotalPrice)
	})
}
