package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
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

	_, err = s.storage.CreateOrder(ctx, order)
	if err != nil {
		return err
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
