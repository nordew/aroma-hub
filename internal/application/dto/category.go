package dto

import "aroma-hub/internal/models"

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type ListCategoryResponse struct {
	Categories []models.Category `json:"categories"`
	Total      int64             `json:"total"`
}

type ListCategoryFilter struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Limit uint   `json:"limit"`
	Page  uint   `json:"page"`
}
