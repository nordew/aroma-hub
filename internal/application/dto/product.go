package dto

import "aroma-hub/internal/models"

type CreateProductRequest struct {
	CategoryName    string  `json:"categoryName"`
	Brand           string  `json:"brand"`
	Name            string  `json:"name"`
	ImageURL        string  `json:"imageUrl"`
	Description     string  `json:"description"`
	Composition     string  `json:"composition"`
	Characteristics string  `json:"characteristics"`
	Price           float64 `json:"price"`
	StockAmount     uint    `json:"stockAmount"`
}

type ListProductResponse struct {
	Products []models.Product `json:"products"`
	Count    int64            `json:"count"`
}

type ListProductFilter struct {
	IDs             []string `json:"id"`
	CategoryID      string   `json:"categoryId"`
	CategoryName    string   `json:"categoryName"`
	Brand           string   `json:"brand"`
	Name            string   `json:"name"`
	PriceFrom       uint     `json:"priceFrom"`
	PriceTo         uint     `json:"priceTo"`
	StockAmountFrom uint     `json:"leftFrom"`
	StockAmountTo   uint     `json:"leftTo"`
	SortBy          string   `json:"sortBy"`
	SortOrder       string   `json:"sortOrder"`
	Limit           uint     `json:"limit"`
	Page            uint     `json:"page"`
}

type UpdateProductRequest struct {
	ID              string  `json:"id"`
	CategoryName    string  `json:"categoryName"`
	Brand           string  `json:"brand"`
	Name            string  `json:"name"`
	ImageURL        string  `json:"imageUrl"`
	Description     string  `json:"description"`
	Composition     string  `json:"composition"`
	Characteristics string  `json:"characteristics"`
	Price           float64 `json:"price"`
	StockAmount     uint    `json:"stockAmount"`
}
