package dto

type AdminLoginRequest struct {
	OTP string `json:"otp" validate:"required"`
}

type AdminLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type AdminRefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type AdminRefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ListAdminFilter struct {
	VendorID string `json:"vendorId"`
}
