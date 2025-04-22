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
	UserID        string               `json:"userId" validate:"required"`
	FullName      string               `json:"fullName" validate:"required"`
	PhoneNumber   string               `json:"phoneNumber" validate:"required"`
	Address       string               `json:"address" validate:"required"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod" validate:"required,oneof=IBAN сash_on_delivery"`
	PromoCode     string               `json:"promoCode"`
	ContactType   models.ContactType   `json:"contactType" validate:"required,oneof=telegram phone"`
	ProductItems  []ProductOrderItem   `json:"items" validate:"required"`
}

type ListOrdersResponse struct {
	Orders []models.Order `json:"orders"`
	Total  int64          `json:"total"`
}

type ListOrderFilter struct {
	Limit uint `json:"limit"`
	Page  uint `json:"page"`

	IDs           string               `json:"id"`
	UserID        string               `json:"userId"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod"`
	ContactType   models.ContactType   `json:"contactType"`
	FromDate      *time.Time           `json:"fromDate"`
	ToDate        *time.Time           `json:"toDate"`
	Status        models.OrderStatus   `json:"status"`
}
