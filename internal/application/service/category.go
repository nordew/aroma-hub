package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"

	"github.com/google/uuid"
)

func (s *Service) CreateCategory(ctx context.Context, input dto.CreateCategoryRequest) error {
	category, err := models.NewCategory(uuid.NewString(), input.Name)
	if err != nil {
		return err
	}

	return s.storage.CreateCategory(ctx, category)
}

func (s *Service) ListCategories(ctx context.Context, filter dto.ListCategoryFilter) (dto.ListCategoryResponse, error) {
	categories, total, err := s.storage.ListCategories(ctx, filter)
	if err != nil {
		return dto.ListCategoryResponse{}, err
	}

	return dto.ListCategoryResponse{
		Categories: categories,
		Total:      total,
	}, nil
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	return s.storage.DeleteCategory(ctx, id)
}
