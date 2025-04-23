package models

import "time"

type Vendor string

const (
	VendorTelegram Vendor = "telegram"
)

type Admin struct {
	ID         string    `json:"id"`
	VendorID   string    `json:"vendorId"`
	VendorType Vendor    `json:"vendorType"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func NewAdmin(id string, vendorID string, vendorType Vendor) (Admin, error) {
	return Admin{
		ID:         id,
		VendorID:   vendorID,
		VendorType: vendorType,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}
