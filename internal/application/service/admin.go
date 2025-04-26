package service

import (
	"aroma-hub/internal/application/dto"
	"context"
	"github.com/nordew/go-errx"
)

func (s *Service) IsAdmin(ctx context.Context, vendorID string) (bool, error) {
	admins, err := s.storage.ListAdmins(ctx, dto.ListAdminFilter{
		VendorID: vendorID,
	})
	if err != nil {
		return false, err
	}

	return len(admins) > 0, nil
}

func (s *Service) AdminLogin(ctx context.Context, input dto.AdminLoginRequest) (dto.AdminLoginResponse, error) {
	vendorID, ok := s.cache.Get(input.OTP)
	if !ok {
		return dto.AdminLoginResponse{}, errx.NewBadRequest().WithDescription("Invalid OTP")
	}

	admins, err := s.storage.ListAdmins(ctx, dto.ListAdminFilter{
		VendorID: vendorID.(string),
	})
	if err != nil {
		return dto.AdminLoginResponse{}, err
	}
	admin := admins[0]

	if admin.VendorID != vendorID {
		return dto.AdminLoginResponse{}, errx.NewBadRequest().WithDescription("Invalid OTP")
	}

	accessToken, err := s.tokenService.GenerateAccessToken(admin.ID, admin.VendorID)
	if err != nil {
		return dto.AdminLoginResponse{}, err
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(admin.ID, admin.VendorID)
	if err != nil {
		return dto.AdminLoginResponse{}, err
	}

	return dto.AdminLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) AdminRefresh(_ context.Context, input dto.AdminRefreshTokenRequest) (dto.AdminRefreshTokenResponse, error) {
	return dto.AdminRefreshTokenResponse{}, nil
	//accessToken, err := s.tokenService.GenerateAccessToken(input.AdminID, input.VendorID)
	//if err != nil {
	//	return dto.AdminLoginResponse{}, err
	//}
	//
	//refreshToken, err := s.tokenService.GenerateRefreshToken(input.AdminID, input.VendorID)
	//if err != nil {
	//	return dto.AdminLoginResponse{}, err
	//}
	//
	//return dto.AdminLoginResponse{
	//	AccessToken:  accessToken,
	//	RefreshToken: refreshToken,
	//}, nil
}
