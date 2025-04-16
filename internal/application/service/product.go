package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"github.com/google/uuid"
	"github.com/nordew/go-errx"
	"time"
)

type productStorage interface {
	Create(ctx context.Context, product models.Product) error
	List(ctx context.Context, filter dto.ListProductFilter) ([]models.Product, int64, error)
	Delete(ctx context.Context, id string) error
}

type categoryStorage interface {
	List(ctx context.Context, filter dto.ListCategoryFilter) ([]models.Category, int64, error)
}

type ProductService struct {
	productStorage  productStorage
	categoryStorage categoryStorage
}

func NewProductService(productStorage productStorage, storage categoryStorage) *ProductService {
	return &ProductService{
		productStorage:  productStorage,
		categoryStorage: storage,
	}
}

func (s *ProductService) Create(ctx context.Context, input dto.CreateProductRequest) error {
	now := time.Now()

	_, _, err := s.productStorage.List(ctx, dto.ListProductFilter{
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

	return s.productStorage.Create(ctx, product)
}

func (s *ProductService) List(ctx context.Context, filter dto.ListProductFilter) (dto.ListProductResponse, error) {
	products, total, err := s.productStorage.List(ctx, filter)
	if err != nil {
		return dto.ListProductResponse{}, err
	}

	return dto.ListProductResponse{
		Products: products,
		Count:    total,
	}, nil
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.productStorage.Delete(ctx, id)
}
