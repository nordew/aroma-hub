package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/nordew/go-errx"
)

var (
	ErrInvalidID         = "invalid category ID"
	ErrEmptyCategoryName = "category name cannot be empty"
)

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewCategory(id string, name string) (Category, error) {
	c := Category{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := c.Validate(); err != nil {
		return Category{}, err
	}

	return c, nil
}

func (c *Category) Validate() error {
	if _, err := uuid.Parse(c.ID); err != nil {
		return errx.NewValidation().WithDescription(ErrInvalidID)
	}

	if c.Name == "" {
		return errx.NewValidation().WithDescription(ErrEmptyCategoryName)
	}

	return nil
}
