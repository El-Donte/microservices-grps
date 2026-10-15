package domain

import (
	"errors"
)

var (
	CartNotFound  error = errors.New("cart not found")
	CartIsEmpty   error = errors.New("cart is empty")
	InvalidUserId error = errors.New("user_id must be a positive integer")
	InvalidSku    error = errors.New("sku must be a positive integer")
	InvalidPath   error = errors.New("invalid path")
	InvalidCount  error = errors.New("count must be a positive integer")
	SkuNotFound   error = errors.New("sku does not exist in product-service")
)
