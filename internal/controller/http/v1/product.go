package v1

import (
	"aroma-hub/internal/application/dto"
	"context"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) initProductRoutes(api fiber.Router) {
	products := api.Group("/products")

	products.Get("/", h.getProducts)
	products.Post("/", h.createProduct)
	products.Delete("/:id", h.deleteProduct)
}

func (h *Handler) getProducts(c *fiber.Ctx) error {
	var filter dto.ListProductFilter
	if err := c.QueryParser(&filter); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, "invalid query params")
	}

	resp, err := h.productService.List(context.Background(), filter)
	if err := handleError(c, err, "listing products"); err != nil {
		return err
	}

	return writeResponse(c, fiber.StatusOK, resp)
}

func (h *Handler) createProduct(c *fiber.Ctx) error {
	var input dto.CreateProductRequest
	if err := c.BodyParser(&input); err != nil {
		return writeErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	err := h.productService.Create(context.Background(), input)
	if err := handleError(c, err, "creating product"); err != nil {
		return err
	}

	return writeResponse(c, fiber.StatusCreated, nil)
}

func (h *Handler) deleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return writeErrorResponse(c, fiber.StatusBadRequest, "invalid product id")
	}

	err := h.productService.Delete(context.Background(), id)
	if err := handleError(c, err, "deleting product"); err != nil {
		return err
	}

	return writeResponse(c, fiber.StatusNoContent, nil)
}
