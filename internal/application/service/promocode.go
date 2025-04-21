package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"log"
)

func (s *Service) CreatePromocode(ctx context.Context, input dto.CreatePromocodeRequest) error {
	log.Printf("Creating promocode with code %s, discount %d%% and expiration %s", input.Code, input.Discount, input.ExpiresAt.Format("2006-01-02"))

	promocode, err := models.NewPromocode(input.Code, input.Discount, input.ExpiresAt)
	if err != nil {
		return err
	}

	return s.storage.CreatePromocode(ctx, promocode)
}

func (s *Service) ListPromocodes(ctx context.Context, filter dto.ListPromocodeFilter) (dto.ListPromocodesResponse, error) {
	promocodes, total, err := s.storage.ListPromocodes(ctx, filter)
	if err != nil {
		return dto.ListPromocodesResponse{}, err
	}

	return dto.ListPromocodesResponse{
		Promocodes: promocodes,
		Total:      total,
	}, nil
}

func (s *Service) DeletePromocode(ctx context.Context, id string) error {
	return s.storage.DeletePromocode(ctx, id)
}

func (s *Service) DeleteExpiredPromocodes(ctx context.Context) (int64, error) {
	promocodes, total, err := s.storage.ListPromocodes(ctx, dto.ListPromocodeFilter{Expired: true})
	if err != nil {
		return 0, err
	}

	for _, promocode := range promocodes {
		err := s.storage.DeletePromocode(ctx, promocode.ID)
		if err != nil {
			return 0, err
		}
	}

	return total, nil
}
