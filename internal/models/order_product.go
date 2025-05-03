package models

import "github.com/nordew/go-errx"

var (
	ErrInvalidQuantity = errx.NewValidation().WithDescription("quantity must be greater than zero")
	ErrInvalidVolume   = errx.NewValidation().WithDescription("volume must be between 2 and 10")
)

type OrderProduct struct {
	OrderID   string `json:"orderId"`
	ProductID string `json:"productId"`
	Quantity  uint   `json:"quantity"`
	Volume    uint   `json:"volume"`
}

func NewOrderProduct(orderID, productID string, quantity, volume uint) (OrderProduct, error) {
	if quantity <= 0 {
		return OrderProduct{}, ErrInvalidQuantity
	}
	if volume < 2 || volume > 10 {
		return OrderProduct{}, ErrInvalidVolume
	}

	return OrderProduct{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Volume:    volume,
	}, nil
}
