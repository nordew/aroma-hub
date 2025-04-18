package models

import (
	"github.com/nordew/go-errx"
	"strings"
	"time"
)

const (
	ErrEmptyID             = "id cannot be empty"
	ErrEmptyCategoryID     = "category id cannot be empty"
	ErrProductEmptyName    = "name cannot be empty"
	ErrInvalidImageURL     = "image URL is invalid"
	ErrInvalidProductPrice = "price must be greater than zero"
)

type Product struct {
	ID              string    `json:"id"`
	CategoryID      string    `json:"-"`
	CategoryName    string    `json:"categoryName"`
	Brand           string    `json:"brand"`
	Name            string    `json:"name"`
	ImageURL        string    `json:"imageUrl"`
	Description     string    `json:"description"`
	Composition     string    `json:"composition"`
	Characteristics string    `json:"characteristics"`
	Price           uint      `json:"price"`
	StockAmount     uint      `json:"stockAmount"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func NewProduct(
	id, categoryID, brand, name, imageURL, description, composition, characteristics string,
	price, stockAmount uint,
) (*Product, error) {
	p := &Product{
		ID:              id,
		CategoryID:      categoryID,
		Brand:           brand,
		Name:            name,
		ImageURL:        imageURL,
		Description:     description,
		Composition:     composition,
		Characteristics: characteristics,
		Price:           price,
		StockAmount:     stockAmount,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Product) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errx.NewValidation().WithDescription(ErrEmptyID)
	}
	if strings.TrimSpace(p.CategoryID) == "" {
		return errx.NewValidation().WithDescription(ErrEmptyCategoryID)
	}
	if strings.TrimSpace(p.Name) == "" {
		return errx.NewValidation().WithDescription(ErrProductEmptyName)
	}
	if p.ImageURL != "" && !strings.HasPrefix(p.ImageURL, "http") {
		return errx.NewValidation().WithDescription(ErrInvalidImageURL)
	}
	if p.Price == 0 {
		return errx.NewValidation().WithDescription(ErrInvalidProductPrice)
	}

	return nil
}
