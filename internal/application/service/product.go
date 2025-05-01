package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/nordew/go-errx"
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
		return dto.ListProductResponse{}, fmt.Errorf("failed to list products: %w", err)
	}

	return dto.ListProductResponse{
		Products: products,
		Count:    total,
	}, nil
}

func (s *Service) UpdateProduct(ctx context.Context, input dto.UpdateProductRequest) error {
	var newCategoryName string
	if input.CategoryName != "" {
		categories, _, err := s.storage.ListCategories(ctx, dto.ListCategoryFilter{
			Name: input.CategoryName,
		})
		if err != nil {
			return err
		}
		category := categories[0]

		newCategoryName = category.Name
	}

	input.CategoryName = newCategoryName

	return s.storage.UpdateProduct(ctx, input)
}

func (s *Service) SetProductImage(ctx context.Context, productID string, imageBytes []byte) error {
	return errx.NewInternal().WithDescription("not implemented")
}

func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	return s.storage.DeleteProduct(ctx, id)
}
