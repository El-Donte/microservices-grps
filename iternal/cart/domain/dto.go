package domain

type AddToCartRequest struct {
	Count uint32 `json:"count"`
}

type CartItemResponse struct {
	SKU   int64  `json:"sku"`
	Name  string `json:"name"`
	Count uint32 `json:"count"`
	Price uint32 `json:"price"`
}

type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalPrice uint32             `json:"total_price"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
