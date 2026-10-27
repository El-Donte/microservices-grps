package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/El-Donte/microservices-grps/docs"
	"github.com/El-Donte/microservices-grps/iternal/cart/client"
	"github.com/El-Donte/microservices-grps/iternal/cart/domain"
	"github.com/El-Donte/microservices-grps/iternal/cart/middleware"
	"github.com/El-Donte/microservices-grps/iternal/cart/repository"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartService struct {
	repo          *repository.Repository
	productClient client.ProductClient
	mux           *http.ServeMux
}

func NewCartService(repo *repository.Repository, productClient client.ProductClient) *CartService {
	service := &CartService{repo: repo, productClient: productClient, mux: http.NewServeMux()}
	service.routes()
	return service
}

func (s *CartService) Handler() http.Handler {
	return middleware.LoggingMiddleware(s.mux)
}

func (s *CartService) routes() {
	s.mux.HandleFunc("/user/", s.handleUserCart)
	s.mux.HandleFunc("/", s.handleRoot)

	s.mux.Handle(
		"/swagger/",
		httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		),
	)

	s.mux.HandleFunc(
		"/swagger/doc.json",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(docs.SwaggerInfo.ReadDoc()))
		})
}

// @Summary Get cart contents
// @Description Returns the current contents of the user's cart. Items are sorted by SKU ascending.
// @Tags cart
// @Produce json
// @Param user_id path int64 true "User ID (must be > 0)"
// @Success 200 {object} domain.CartResponse "Cart contents retrieved successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid user_id (must be > 0)"
// @Failure 404 {object} domain.ErrorResponse "Cart not found or is empty"
// @Router /user/{user_id}/cart [get]
func (s *CartService) handleUserCart(writer http.ResponseWriter, request *http.Request) {
	userId, sku, err := parseCartPath(request.URL.Path)
	if err != nil {
		if errors.Is(err, domain.InvalidPath) {
			writeError(writer, http.StatusInternalServerError, err.Error())
			return
		}
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}

	switch {
	case request.Method == http.MethodPost:
		s.addToCart(writer, request, userId, sku)
	case request.Method == http.MethodGet:
		s.getCart(writer, request, userId)
	default:
		writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
	}
}
func (s *CartService) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"service": "cart",
		"swagger": "/swagger/",
	})
}

// @Summary Add item to cart
// @Description Adds a specified item to a specific user's cart. Quantities are summed if the item already exists.
// @Tags cart
// @Accept json
// @Produce json
// @Param user_id path int64 true "User ID (must be > 0)"
// @Param sku_id path int64 true "Product SKU ID (must be > 0)"
// @Param request body domain.AddToCartRequest true "Item count to add"
// @Success 200 "Item added successfully"
// @Failure 400 {object} domain.ErrorResponse "Invalid user_id, sku_id, or count (must be > 0)"
// @Failure 412 {object} domain.ErrorResponse "SKU does not exist in product-service"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /user/{user_id}/cart/{sku_id} [post]
func (s *CartService) addToCart(writer http.ResponseWriter, request *http.Request, userId uint64, sku uint64) {
	var body = domain.AddToCartRequest{}

	if request.Body != nil {
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			writeError(writer, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	if body.Count <= 0 {
		writeError(writer, http.StatusBadRequest, domain.InvalidCount.Error())
		return
	}

	name, price, err := s.productClient.GetProduct(request.Context(), sku)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			writeError(writer, http.StatusPreconditionFailed,
				domain.SkuNotFound.Error())
			return
		}
		if ok && st.Code() == codes.InvalidArgument {
			writeError(writer, http.StatusBadRequest, st.Message())
			return
		}

		writeError(writer, http.StatusInternalServerError, "product-service unavailable")
		return
	}

	cart, err := s.repo.Get(userId)
	if err != nil {
		cart = s.repo.Create(userId)
	}

	cart.AddItem(sku, name, body.Count, price)

	writer.WriteHeader(http.StatusOK)
}

func (s *CartService) getCart(writer http.ResponseWriter, request *http.Request, userId uint64) {
	cart, err := s.repo.Get(userId)
	if err != nil || cart == nil {
		writeError(writer, http.StatusNotFound, domain.CartNotFound.Error())
		return
	}

	if cart.IsEmpty() {
		writeError(writer, http.StatusNotFound, domain.CartIsEmpty.Error())
		return
	}

	items := cart.Items()

	skus := make([]uint64, 0, len(items))
	counts := make(map[uint64]int32, len(items))
	for _, it := range items {
		skus = append(skus, it.SKU)
		counts[it.SKU] = int32(it.Count)
	}

	ctx, cancel := context.WithTimeout(request.Context(), 10*time.Second)
	defer cancel()

	products, err := s.productClient.FetchProducts(ctx, skus)

	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			writeError(writer, http.StatusPreconditionFailed, "product no longer available")
			return
		}

		log.Printf("cart: FetchAll(%d): %v", userId, err)
		writeError(writer, http.StatusInternalServerError, "product-service unavailable")
		return
	}

	type itemResponse struct {
		SKU   int64  `json:"sku"`
		Name  string `json:"name"`
		Count int32  `json:"count"`
		Price int32  `json:"price"`
	}

	respItems := make([]itemResponse, 0, len(products))
	var total uint64
	for _, it := range products {
		respItems = append(respItems, itemResponse{
			SKU:   int64(it.SKU),
			Name:  it.Name,
			Count: counts[it.SKU],
			Price: int32(it.Price),
		})
		total += uint64(it.Price) * uint64(counts[it.SKU])
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(map[string]any{
		"items":       respItems,
		"total_price": int32(total),
	})
}

func parseCartPath(path string) (userId, sku uint64, err error) {
	const prefix = "/user/"
	if !strings.HasPrefix(path, prefix) {
		return 0, 0, domain.InvalidPath
	}

	parts := strings.Split(strings.Trim(strings.TrimPrefix(path, prefix), "/"), "/")
	if len(parts) < 2 || parts[1] != "cart" {
		return 0, 0, domain.InvalidPath
	}

	uid, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || uid == 0 {
		return 0, 0, domain.InvalidUserId
	}
	userId = uid

	switch len(parts) {
	case 2:
		return userId, 0, nil
	case 3:
		sku, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil || sku == 0 {
			return 0, 0, domain.InvalidSku
		}
		return userId, sku, nil
	default:
		return 0, 0, domain.InvalidPath
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
