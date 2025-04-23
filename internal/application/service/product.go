package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (s *Service) CreateProduct(ctx context.Context, input dto.CreateProductRequest) error {
	categories, _, err := s.storage.ListCategories(ctx, dto.ListCategoryFilter{
		Name: input.CategoryName,
	})
	if err != nil {
		return err
	}
	category := categories[0]

	product, err := models.NewProduct(
		uuid.NewString(),
		category.ID,
		input.Brand,
		input.Name,
		input.ImageURL,
		input.Description,
		input.Composition,
		input.Characteristics,
		decimal.NewFromFloat(input.Price),
		input.StockAmount,
	)
	if err != nil {
		return err
	}

	return s.storage.CreateProduct(ctx, product)
}

func (s *Service) ListProducts(ctx context.Context, filter dto.ListProductFilter) (dto.ListProductResponse, error) {
	products, total, err := s.storage.ListProducts(ctx, filter)
	if err != nil {
		return dto.ListProductResponse{}, err
	}

	return dto.ListProductResponse{
		Products: products,
		Count:    total,
	}, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	return s.storage.DeleteProduct(ctx, id)
}
