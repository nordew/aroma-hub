package dto

import (
	"aroma-hub/internal/models"
	"time"
)

type ProductOrderItem struct {
	ID       string `json:"id"`
	Quantity uint   `json:"quantity"`
}

type CreateOrderRequest struct {
	FullName      string               `json:"fullName" validate:"required"`
	PhoneNumber   string               `json:"phoneNumber" validate:"required"`
	Address       string               `json:"address" validate:"required"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod" validate:"required,oneof=IBAN сash_on_delivery"`
	PromoCode     string               `json:"promoCode"`
	ContactType   models.ContactType   `json:"contactType" validate:"required,oneof=telegram phone"`
	ProductItems  []ProductOrderItem   `json:"productItems" validate:"required"`
}

type UpdateOrderRequest struct {
	ID            string               `json:"id" validate:"required"`
	FullName      string               `json:"fullName,omitempty"`
	PhoneNumber   string               `json:"phoneNumber,omitempty"`
	Address       string               `json:"address,omitempty"`
	Status        models.OrderStatus   `json:"status,omitempty"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod,omitempty"`
}

type ListOrdersResponse struct {
	Orders []models.Order `json:"orders"`
	Total  int64          `json:"total"`
}

type ListOrderFilter struct {
	Limit uint `json:"limit"`
	Page  uint `json:"page"`

	IDs           []string             `json:"id"`
	UserID        string               `json:"userId"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod"`
	ContactType   models.ContactType   `json:"contactType"`
	FromDate      *time.Time           `json:"fromDate"`
	ToDate        *time.Time           `json:"toDate"`
	Status        models.OrderStatus   `json:"status"`
}
