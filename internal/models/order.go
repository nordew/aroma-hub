package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/nordew/go-errx"
)

var (
	ErrIDRequired            = "ID is required"
	ErrUserIDRequired        = "UserID is required"
	ErrFullNameRequired      = "FullName is required"
	ErrPhoneNumberInvalid    = "PhoneNumber is invalid"
	ErrAddressRequired       = "Address is required"
	ErrPaymentMethodRequired = "PaymentMethod is required"
	ErrPromoCodeRequired     = "PromoCode is required"
	ErrContactTypeRequired   = "ContactType is required"
	ErrAmountToPayInvalid    = "AmountToPay must be greater than 0"
)

const (
	RegexUkrainianPhone = `^(\+?38)?(0\d{9})$`
)

type PaymentMethod string

const (
	PaymentMethodIBAN           PaymentMethod = "IBAN"
	PaymentMethodCashOnDelivery PaymentMethod = "сash_on_delivery"
)

type ContactType string

const (
	ContactTypeTelegram ContactType = "telegram"
	ContactTypePhone    ContactType = "phone"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
)

type Order struct {
	ID            string        `json:"id"`
	UserID        string        `json:"userId"`
	FullName      string        `json:"fullName"`
	PhoneNumber   string        `json:"phoneNumber"`
	Address       string        `json:"address"`
	PaymentMethod PaymentMethod `json:"paymentMethod"`
	PromoCode     string        `json:"promoCode"`
	ContactType   ContactType   `json:"contactType"`
	AmountToPay   float64       `json:"amountToPay"`
	Status        OrderStatus   `json:"status"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

func NewOrder(
	userID string,
	fullName string,
	phoneNumber string,
	address string,
	paymentMethod PaymentMethod,
	promoCode string,
	contactType ContactType,
	amountToPay float64) (Order, error) {
	now := time.Now()

	order := Order{
		ID:            uuid.New().String(),
		UserID:        userID,
		FullName:      fullName,
		PhoneNumber:   phoneNumber,
		Address:       address,
		PaymentMethod: paymentMethod,
		PromoCode:     promoCode,
		ContactType:   contactType,
		AmountToPay:   amountToPay,
		Status:        OrderStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := order.validate(); err != nil {
		return Order{}, err
	}

	return order, nil
}

func (o Order) validate() error {
	if o.ID == "" {
		return errx.NewValidation().WithDescription(ErrIDRequired)
	}
	if o.UserID == "" {
		return errx.NewValidation().WithDescription(ErrUserIDRequired)
	}
	if o.FullName == "" {
		return errx.NewValidation().WithDescription(ErrFullNameRequired)
	}
	if o.PhoneNumber == "" || !regexp.MustCompile(RegexUkrainianPhone).MatchString(o.PhoneNumber) {
		return errx.NewValidation().WithDescription(ErrPhoneNumberInvalid)
	}
	if o.Address == "" {
		return errx.NewValidation().WithDescription(ErrAddressRequired)
	}
	if o.PaymentMethod == "" {
		return errx.NewValidation().WithDescription(ErrPaymentMethodRequired)
	}
	if o.PromoCode == "" {
		return errx.NewValidation().WithDescription(ErrPromoCodeRequired)
	}
	if o.ContactType == "" {
		return errx.NewValidation().WithDescription(ErrContactTypeRequired)
	}
	if o.AmountToPay <= 0 {
		return errx.NewValidation().WithDescription(ErrAmountToPayInvalid)
	}
	return nil
}

func (pm *PaymentMethod) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*pm = PaymentMethod(v)
	case []byte:
		*pm = PaymentMethod(string(v))
	default:
		return fmt.Errorf("unsupported type %T", value)
	}
	return nil
}

func (ct *ContactType) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*ct = ContactType(v)
	case []byte:
		*ct = ContactType(string(v))
	default:
		return fmt.Errorf("unsupported type %T", value)
	}
	return nil
}
