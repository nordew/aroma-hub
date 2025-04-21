package models

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/nordew/go-errx"
)

var (
	ErrInvalidPromocodeLength     = "code cannot be empty or less than 3 characters or more than 10 characters"
	ErrInvalidPromocodeDiscount   = "discount cannot be greater than 100 or less than 0"
	ErrInvalidPromocodeExpiration = "expiration cannot be in the past"
)

type Promocode struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Discount  uint      `json:"discount"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewPromocode(code string, discount uint, expiresAt time.Time) (Promocode, error) {
	if code == "" || utf8.RuneCountInString(code) < 3 || utf8.RuneCountInString(code) > 10 {
		return Promocode{}, errx.NewValidation().WithDescription(ErrInvalidPromocodeLength)
	}
	if discount <= 0 || discount >= 100 {
		return Promocode{}, errx.NewValidation().WithDescription(ErrInvalidPromocodeDiscount)
	}
	if expiresAt.Before(time.Now()) {
		return Promocode{}, errx.NewValidation().WithDescription(ErrInvalidPromocodeExpiration)
	}

	return Promocode{
		ID:        uuid.NewString(),
		Code:      code,
		Discount:  discount,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
