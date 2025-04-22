package v1

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Middleware struct {
	logger *slog.Logger
}

func NewMiddleware(logger *slog.Logger) *Middleware {
	return &Middleware{logger: logger}
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
