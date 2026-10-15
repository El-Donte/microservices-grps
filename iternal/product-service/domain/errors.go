package domain

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	SkuMustBePositive    error = status.Error(codes.InvalidArgument, "sku must be positive")
	ProductNotFoundError error = status.Error(codes.NotFound, "product not found")
)
