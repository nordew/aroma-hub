package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"time"

	"github.com/nordew/go-errx"
)

func (s *Service) CreateOrder(ctx context.Context, input dto.CreateOrderRequest) error {
	order, err := models.NewOrder(
		input.UserID,
		input.FullName,
		input.PhoneNumber,
		input.Address,
		input.PaymentMethod,
		input.PromoCode,
		input.ContactType,
		input.AmountToPay,
	)
	if err != nil {
		return err
	}

	if order.PromoCode != "" {
		if err := s.checkPromoCode(ctx, order.PromoCode); err != nil {
			return err
		}
	}

	_, err = s.storage.CreateOrder(ctx, order)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) checkPromoCode(ctx context.Context, promoCode string) error {
	promoCodes, _, err := s.storage.ListPromocodes(ctx, dto.ListPromocodeFilter{
		Code: promoCode,
	})
	if err != nil {
		return err
	}
	code := promoCodes[0]

	if code.ExpiresAt.Before(time.Now()) {
		return errx.NewForbidden().WithDescription("Promo code has expired")
	}

	return nil
}

func (s *Service) ListOrders(ctx context.Context, filter dto.ListOrderFilter) (dto.ListOrdersResponse, error) {
	orders, total, err := s.storage.ListOrders(ctx, filter)
	if err != nil {
		return dto.ListOrdersResponse{}, err
	}

	return dto.ListOrdersResponse{
		Orders: orders,
		Total:  total,
	}, nil
}

func (s *Service) DeleteOrder(ctx context.Context, id string) error {
	err := s.storage.DeleteOrder(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
