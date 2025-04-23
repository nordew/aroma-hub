package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Issuer = "Aroma"
)

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrExpiredToken    = errors.New("token has expired")
	ErrTokenGeneration = errors.New("failed to generate token")
	ErrInvalidClaims   = errors.New("invalid token claims")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID    string    `json:"userId"`
	VendorID  string    `json:"vendorId,omitempty"`
	TokenType TokenType `json:"tokenType"`
}

type Config struct {
	AccessTokenSecret    string
	RefreshTokenSecret   string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

func DefaultConfig() Config {
	return Config{
		AccessTokenSecret:    "default_access_token_secret",
		RefreshTokenSecret:   "default_refresh_token_secret",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 24 * time.Hour * 7,
		Issuer:               Issuer,
	}
}

type TokenService struct {
	config Config
}

func NewTokenService(config Config) *TokenService {
	return &TokenService{
		config: config,
	}
}

func NewDefaultTokenService() *TokenService {
	return NewTokenService(DefaultConfig())
}

func (s *TokenService) GenerateAccessToken(userID, vendorID string) (string, error) {
	return s.generateToken(userID, vendorID, AccessToken, s.config.AccessTokenSecret, s.config.AccessTokenDuration)
}

func (s *TokenService) GenerateRefreshToken(userID, vendorID string) (string, error) {
	return s.generateToken(userID, vendorID, RefreshToken, s.config.RefreshTokenSecret, s.config.RefreshTokenDuration)
}

func (s *TokenService) generateToken(userID, vendorID string, tokenType TokenType, secret string, duration time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   userID,
		},
		UserID:    userID,
		VendorID:  vendorID,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	return signedToken, nil
}

func (s *TokenService) VerifyAccessToken(tokenString string) (*Claims, error) {
	return s.verifyToken(tokenString, s.config.AccessTokenSecret)
}

func (s *TokenService) VerifyRefreshToken(tokenString string) (*Claims, error) {
	return s.verifyToken(tokenString, s.config.RefreshTokenSecret)
}

func (s *TokenService) verifyToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method: %v", ErrInvalidToken, token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

func (s *TokenService) RefreshTokens(refreshToken string) (accessToken string, newRefreshToken string, err error) {
	claims, err := s.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	if claims.TokenType != RefreshToken {
		return "", "", fmt.Errorf("%w: not a refresh token", ErrInvalidToken)
	}

	accessToken, err = s.GenerateAccessToken(claims.UserID, claims.VendorID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err = s.GenerateRefreshToken(claims.UserID, claims.VendorID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
