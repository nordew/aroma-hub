package service

import (
	"aroma-hub/internal/application/dto"
	"context"
)

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
