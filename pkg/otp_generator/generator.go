package otp_generator

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	DefaultIssuer = "Aroma"
)

var (
	ErrGeneratingOTP = errors.New("error generating OTP code")
	ErrValidatingOTP = errors.New("error validating OTP code")
	ErrInvalidOTP    = errors.New("invalid OTP code")
	ErrExpiredOTP    = errors.New("OTP code has expired")
)

type Config struct {
	Digits         otp.Digits
	ValidityPeriod uint
	Issuer         string
	Secret         string
}

func DefaultConfig() Config {
	return Config{
		Digits:         otp.DigitsSix,
		ValidityPeriod: 300,
		Issuer:         DefaultIssuer,
	}
}

type Generator struct {
	config Config
}

func NewGenerator(config Config) *Generator {
	return &Generator{
		config: config,
	}
}

func NewDefaultGenerator() *Generator {
	return NewGenerator(DefaultConfig())
}

func (g *Generator) GenerateOTP(accountName string) (string, error) {
	secret := g.config.Secret
	if secret == "" {
		var err error
		secret, err = generateSecureSecret()
		if err != nil {
			return "", fmt.Errorf("%w: failed to generate secure secret: %v", ErrGeneratingOTP, err)
		}
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      g.config.Issuer,
		AccountName: accountName,
		SecretSize:  20,
		Digits:      g.config.Digits,
		Period:      g.config.ValidityPeriod,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGeneratingOTP, err)
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGeneratingOTP, err)
	}

	return code, nil
}

func (g *Generator) GenerateNumericOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGeneratingOTP, err)
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (g *Generator) ValidateOTP(secret, code string) error {
	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now(),
		totp.ValidateOpts{
			Digits:    g.config.Digits,
			Period:    g.config.ValidityPeriod,
			Skew:      1,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrValidatingOTP, err)
	}

	if !valid {
		return ErrInvalidOTP
	}

	return nil
}

func generateSecureSecret() (string, error) {
	b := make([]byte, 20)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}
