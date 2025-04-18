package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
)

type Storage interface {
	CreateProduct(ctx context.Context, product models.Product) error
	ListProducts(ctx context.Context, filter dto.ListProductFilter) ([]models.Product, int64, error)
	DeleteProduct(ctx context.Context, id string) error

	ListCategories(ctx context.Context, filter dto.ListCategoryFilter) ([]models.Category, int64, error)
}

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}
