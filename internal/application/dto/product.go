package dto

import "aroma-hub/internal/models"

type CreateProductRequest struct {
	CategoryName    string  `json:"categoryName"`
	Brand           string  `json:"brand"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Composition     string  `json:"composition"`
	Characteristics string  `json:"characteristics"`
	Price           float64 `json:"price"`
	IsBestSeller    bool    `json:"isBestSeller"`
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
	StockAmount     uint     `json:"stockAmount"`
	SortBy          string   `json:"sortBy"`
	SortOrder       string   `json:"sortOrder"`
	OnlyBestSellers bool     `json:"onlyBestSellers"`
	ShowInvisible   bool     `json:"-"`
	Limit           uint     `json:"limit"`
	Page            uint     `json:"page"`
}

type BrandResponse struct {
	Brands []string `json:"brands"`
}

type UpdateProductRequest struct {
	ID              string  `json:"-"`
	Image           []byte  `json:"-"`
	CategoryName    string  `json:"categoryName"`
	Brand           string  `json:"brand"`
	Name            string  `json:"name"`
	ImageURL        string  `json:"imageUrl"`
	Description     string  `json:"description"`
	Composition     string  `json:"composition"`
	Characteristics string  `json:"characteristics"`
	Price           float64 `json:"price"`
	StockAmount     uint    `json:"stockAmount"`
	MakeVisible     bool    `json:"makeVisible"`
	Hide            bool    `json:"hide"`
	SetBestSeller   bool    `json:"setBestSeller"`
	UnsetBestSeller bool    `json:"unsetBestSeller"`
}

type SetProductImageRequest struct {
	Image []byte `json:"-"`
}
