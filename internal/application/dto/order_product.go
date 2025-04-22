package dto

type ListOrderProductFilter struct {
	Page  uint `json:"page"`
	Limit uint `json:"limit"`

	OrderIDs   []string `json:"orderIds"`
	ProductIDs []string `json:"productIds"`
}
