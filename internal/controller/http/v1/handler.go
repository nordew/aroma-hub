package v1

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/config"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/nordew/go-errx"
	"log"
)

type productService interface {
	Create(ctx context.Context, input dto.CreateProductRequest) error
	List(ctx context.Context, filter dto.ListProductFilter) (dto.ListProductResponse, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	productService productService
}

func NewHandler(productService productService) *Handler {
	return &Handler{
		productService: productService,
	}
}

func (h *Handler) MustInitAndRun(router *fiber.App, cfg config.Server) {
	api := router.Group(cfg.BasePath)

	h.initProductRoutes(api)

	port := fmt.Sprintf(":%d", cfg.Port)
	if err := router.Listen(port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func handleError(c *fiber.Ctx, err error, operation string) error {
	if err == nil {
		return nil
	}

	switch {
	case errx.IsCode(err, errx.NotFound):
		return writeErrorResponse(c, fiber.StatusNotFound, "resource not found")
	case errx.IsCode(err, errx.Internal):
		return writeErrorResponse(c, fiber.StatusInternalServerError, "internal server error: "+operation)
	case errx.IsCode(err, errx.BadRequest):
		return writeErrorResponse(c, fiber.StatusBadRequest, err.Error())
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
