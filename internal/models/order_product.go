package models

import "github.com/nordew/go-errx"

type OrderProduct struct {
	OrderID   string `json:"orderId"`
	ProductID string `json:"productId"`
	Quantity  uint   `json:"quantity"`
}

func NewOrderProduct(orderID, productID string, quantity uint) (OrderProduct, error) {
	if quantity <= 0 {
		return OrderProduct{}, errx.NewValidation().WithDescription("quantity must be greater than zero")
	}

	return OrderProduct{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
	}, nil
}
