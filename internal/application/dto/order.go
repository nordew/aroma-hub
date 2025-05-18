package dto

import (
	"aroma-hub/internal/models"
	"time"
)

type ProductOrder struct {
	ID       string `json:"id" validate:"required"`
	Brand    string `json:"brand" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Price    uint   `json:"price" validate:"required"`
	Quantity uint   `json:"quantity" validate:"required"`
	Volume   uint   `json:"volume" validate:"required"`
}

type Order struct {
	ID            string               `json:"id"`
	FullName      string               `json:"fullName"`
	PhoneNumber   string               `json:"phoneNumber"`
	Address       string               `json:"address"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod"`
	ContactType   models.ContactType   `json:"contactType"`
	AmountToPay   uint                 `json:"amountToPay"`
	Status        models.OrderStatus   `json:"status"`
	Products      []ProductOrder       `json:"products"`
	CreatedAt     time.Time            `json:"createdAt"`
	UpdatedAt     time.Time            `json:"updatedAt"`
}

type OrderResponse struct {
	Orders []Order `json:"orders"`
	Count  uint    `json:"count"`
}

type CreateOrderRequest struct {
	FullName      string               `json:"fullName" validate:"required"`
	PhoneNumber   string               `json:"phoneNumber" validate:"required"`
	Address       string               `json:"address" validate:"required"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod" validate:"required,oneof=IBAN сash_on_delivery"`
	PromoCode     string               `json:"promoCode"`
	ContactType   models.ContactType   `json:"contactType" validate:"required,oneof=telegram phone"`
	ProductItems  []ProductOrder       `json:"productItems" validate:"required"`
}

type UpdateOrderRequest struct {
	ID            string               `json:"id" validate:"required"`
	FullName      string               `json:"fullName,omitempty"`
	PhoneNumber   string               `json:"phoneNumber,omitempty"`
	Address       string               `json:"address,omitempty"`
	Status        models.OrderStatus   `json:"status,omitempty"`
	PaymentMethod models.PaymentMethod `json:"paymentMethod,omitempty"`
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
