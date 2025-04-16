package models

import (
	"github.com/nordew/go-errx"
	"time"
)

var (
	ErrInvalidID         = "invalid category ID"
	ErrEmptyCategoryName = "category name cannot be empty"
)

type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewCategory(id int, name string) (*Category, error) {
	c := &Category{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Category) Validate() error {
	if c.ID <= 0 {
		return errx.NewValidation().WithDescription(ErrInvalidID)
	}
	if c.Name == "" {
		return errx.NewValidation().WithDescription(ErrEmptyCategoryName)
	}

	return nil
}
