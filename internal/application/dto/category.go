package dto

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type ListCategoryFilter struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Limit uint   `json:"limit"`
	Page  uint   `json:"page"`
}
