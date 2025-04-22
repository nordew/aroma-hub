package v1

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/config"
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/labstack/gommon/log"
	"github.com/nordew/go-errx"
)

type Service interface {
	CreateProduct(ctx context.Context, input dto.CreateProductRequest) error
	ListProducts(ctx context.Context, filter dto.ListProductFilter) (dto.ListProductResponse, error)
	DeleteProduct(ctx context.Context, id string) error

	ListCategories(ctx context.Context, filter dto.ListCategoryFilter) (dto.ListCategoryResponse, error)

	CreateOrder(ctx context.Context, order dto.CreateOrderRequest) error
	ListOrders(ctx context.Context, filter dto.ListOrderFilter) (dto.ListOrdersResponse, error)
	DeleteOrder(ctx context.Context, id string) error

	CreatePromocode(ctx context.Context, input dto.CreatePromocodeRequest) error
	ListPromocodes(ctx context.Context, filter dto.ListPromocodeFilter) (dto.ListPromocodesResponse, error)
	DeletePromocode(ctx context.Context, id string) error
}

type Handler struct {
	service    Service
	middleware *Middleware
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service:    service,
		middleware: NewMiddleware(logger),
	}
}

func (h *Handler) InitAndServe(router *fiber.App, cfg config.Server) error {
	router.Use(h.middleware.RequestLogger())

	api := router.Group(cfg.BasePath)

	h.initProductRoutes(api)
	h.initCategoryRoutes(api)
	h.initOrderRoutes(api)
	h.initPromocodeRoutes(api)
	api.Get("/health", h.healthCheck)

	port := fmt.Sprintf(":%d", cfg.Port)
	h.middleware.logger.Info("starting server",
		slog.String("port", port),
		slog.String("basePath", cfg.BasePath),
	)

	return router.Listen(port)
}

func (h *Handler) healthCheck(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func handleError(c *fiber.Ctx, err error, operation string) error {
	if err == nil {
		return nil
	}

	log.Error("error occurred during operation", slog.String("operation", operation), slog.String("error", err.Error()))

	switch {
	case errx.IsCode(err, errx.NotFound):
		return writeErrorResponse(c, fiber.StatusNotFound, err.Error())
	case errx.IsCode(err, errx.Internal):
		return writeErrorResponse(c, fiber.StatusInternalServerError, "internal server error: "+operation)
	case errx.IsCode(err, errx.BadRequest):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	case errx.IsCode(err, errx.BadRequest):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	case errx.IsCode(err, errx.Validation):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
	case errx.IsCode(err, errx.Unauthorized):
		return writeErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	case errx.IsCode(err, errx.Forbidden):
		return writeErrorResponse(c, fiber.StatusForbidden, err.Error())
	default:
		return writeErrorResponse(c, fiber.StatusInternalServerError, "unexpected error: "+operation)
	}
}

func writeErrorResponse(c *fiber.Ctx, status int, message string) error {
	response := fiber.Map{
		"success": false,
		"message": message,
	}

	return c.Status(status).JSON(response)
}

func writeResponse(c *fiber.Ctx, status int, data interface{}) error {
	response := fiber.Map{
		"success": true,
		"data":    data,
	}

	return c.Status(status).JSON(response)
}
