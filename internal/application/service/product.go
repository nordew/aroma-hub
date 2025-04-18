package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"github.com/google/uuid"
	"github.com/nordew/go-errx"
	"time"
)

func (s *Service) CreateProduct(ctx context.Context, input dto.CreateProductRequest) error {
	now := time.Now()

	_, _, err := s.storage.ListProducts(ctx, dto.ListProductFilter{
		CategoryID: input.CategoryID,
	})
	if err != nil {
		if errx.IsCode(err, errx.NotFound) {
			return errx.NewBadRequest().WithDescription("category not found")
		}

		return errx.NewInternal().WithDescriptionAndCause("failed to list products", err)
	}

	product := models.Product{
		ID:              uuid.NewString(),
		CategoryID:      input.CategoryID,
		Brand:           input.Brand,
		Name:            input.Name,
		ImageURL:        input.ImageURL,
		Description:     input.Description,
		Composition:     input.Composition,
		Characteristics: input.Characteristics,
		Price:           input.Price,
		StockAmount:     input.StockAmount,
		CreatedAt:       now,
		UpdatedAt:       now,
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
