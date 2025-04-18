package v1

import (
	"aroma-hub/internal/application/dto"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/nordew/go-errx"
)

func (h *Handler) initProductRoutes(api fiber.Router) {
	products := api.Group("/products")

	products.Get("/", h.listProducts)
	products.Post("/", h.createProduct)
	products.Delete("/:id", h.deleteProduct)
}

func (h *Handler) listProducts(c *fiber.Ctx) error {
	const op = "listProducts"

	var filter dto.ListProductFilter
	if err := c.QueryParser(&filter); err != nil {
		return handleError(c, err, op)
	}

	resp, err := h.service.ListProducts(context.Background(), filter)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusOK, resp)
}

func (h *Handler) createProduct(c *fiber.Ctx) error {
	const op = "createProduct"

	var input dto.CreateProductRequest
	if err := c.BodyParser(&input); err != nil {
		return handleError(c, err, op)
	}

	err := h.service.CreateProduct(context.Background(), input)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusCreated, nil)
}

func (h *Handler) deleteProduct(c *fiber.Ctx) error {
	const op = "deleteProduct"

	id := c.Params("id")
	if id == "" {
		return handleError(c, errx.NewBadRequest().WithDescription("id is empty"), op)
	}

	err := h.service.DeleteProduct(context.Background(), id)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusNoContent, nil)
}
