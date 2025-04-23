package v1

import (
	"aroma-hub/pkg/auth"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Middleware struct {
	logger       *slog.Logger
	tokenService *auth.TokenService
}

func NewMiddleware(logger *slog.Logger, tokenService *auth.TokenService) *Middleware {
	return &Middleware{
		logger:       logger,
		tokenService: tokenService,
	}
}

func (m *Middleware) Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token")
		}

		splitToken := strings.Split(token, "Bearer ")
		if len(splitToken) != 2 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token format")
		}

		token = strings.TrimSpace(splitToken[1])

		userID, err := m.tokenService.VerifyAccessToken(token)
		if err != nil {
			if errors.Is(err, auth.ErrExpiredToken) {
				return fiber.NewError(fiber.StatusUnauthorized, "token expired")
			}
			if errors.Is(err, auth.ErrInvalidToken) {
				return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
			}

			return fiber.NewError(fiber.StatusInternalServerError, "failed to verify token")
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}

func (m *Middleware) RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		c.Locals("requestID", requestID)
		c.Locals("startTime", start)

		m.logRequestStart(c, requestID)

		err := c.Next()

		m.logRequestEnd(c, requestID, start, err)

		return err
	}
}

func (m *Middleware) logRequestStart(c *fiber.Ctx, requestID string) {
	m.logger.Info("request started",
		slog.String("method", c.Method()),
		slog.String("path", c.Path()),
		slog.String("ip", c.IP()),
		slog.String("requestID", requestID),
		slog.String("userAgent", c.Get("User-Agent")),
	)
}

func (m *Middleware) logRequestEnd(c *fiber.Ctx, requestID string, start time.Time, err error) {
	latency := time.Since(start)
	status := c.Response().StatusCode()
	msg := "request completed"

	attrs := []any{
		slog.String("method", c.Method()),
		slog.String("path", c.Path()),
		slog.Int("status", status),
		slog.Duration("latency", latency),
		slog.String("requestID", requestID),
	}

	if err != nil {
		msg = "request failed"
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	m.logger.Info(msg, attrs...)
}
