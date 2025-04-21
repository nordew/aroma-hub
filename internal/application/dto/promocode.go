package dto

import (
	"aroma-hub/internal/models"
	"time"
)

type CreatePromocodeRequest struct {
	Code      string    `json:"code"`
	Discount  uint      `json:"discount"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type ListPromocodesResponse struct {
	Promocodes []models.Promocode `json:"promocodes"`
	Total      int64              `json:"total"`
}

type ListPromocodeFilter struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	DiscountFrom uint   `json:"discountFrom"`
	DiscountTo   uint   `json:"discountTo"`
	Active       bool   `json:"active"`
	Expired      bool   `json:"expired"`
	Limit        uint   `json:"limit"`
	Page         uint   `json:"page"`
}
