package unit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/El-Donte/microservices-grps/iternal/cart/repository"
	"github.com/El-Donte/microservices-grps/iternal/cart/service"
	"github.com/El-Donte/microservices-grps/iternal/cart/unit-test"
	"github.com/stretchr/testify/require"
)

func newCartService(t *testing.T) *service.CartService {
	t.Helper()

	productService := service.NewCartService(repository.NewRepository(), unit.NewFakeProductService())

	return productService
}

func TestGetCart(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"cart not found", "/user/1/cart", http.StatusNotFound},
		{"user id must be positive", "/user/0/cart", http.StatusBadRequest},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svc := newCartService(t)

			req := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			rr := httptest.NewRecorder()
			svc.Handler().ServeHTTP(rr, req)

			require.Equal(t, testCase.wantStatus, rr.Code)
		})
	}

	t.Run("items sorted and right structure", func(t *testing.T) {
		t.Parallel()
		svc := newCartService(t)

		for _, add := range []struct {
			path, body string
		}{
			{"/user/7/cart/1", `{"count":2}`},
			{"/user/7/cart/2", `{"count":1}`},
			{"/user/7/cart/1", `{"count":1}`},
		} {
			rr := httptest.NewRecorder()
			svc.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodPost, add.path, strings.NewReader(add.body)))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		rr := httptest.NewRecorder()
		svc.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/user/7/cart", nil))

		require.Equal(t, http.StatusOK, rr.Code)

		var body struct {
			Items []struct {
				SKU   int64  `json:"sku"`
				Name  string `json:"name"`
				Count uint32 `json:"count"`
				Price uint32 `json:"price"`
			} `json:"items"`
			TotalPrice uint32 `json:"total_price"`
		}

		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
		require.Len(t, body.Items, 2)
		require.Equal(t, int64(1), body.Items[0].SKU)
		require.Equal(t, uint32(3), body.Items[0].Count)
		require.Equal(t, int64(2), body.Items[1].SKU)
		require.Equal(t, uint32(1), body.Items[1].Count)
		require.Equal(t, uint32(500), body.TotalPrice)
	})
}

func TestAddToCart(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"sku not found", "/user/1/cart/10", `{"count": 3}`, http.StatusPreconditionFailed},
		{"sku must be positive", "/user/1/cart/0", `{"count": 3}`, http.StatusBadRequest},

		{"user id must be positive", "/user/0/cart/1", `{"count": 1}`, http.StatusBadRequest},

		{"body must be positive", "/user/1/cart/1", `{"count": 0}`, http.StatusBadRequest},
		{"invalid body", "/user/1/cart/1", `{"count": "x"}`, http.StatusBadRequest},

		{"happy path", "/user/1/cart/1", `{"count": 3}`, http.StatusOK},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			svc := newCartService(t)
			req := httptest.NewRequest(http.MethodPost, testCase.path, strings.NewReader(testCase.body))

			rr := httptest.NewRecorder()
			svc.Handler().ServeHTTP(rr, req)

			require.Equal(t, testCase.wantStatus, rr.Code)
		})
	}
}
