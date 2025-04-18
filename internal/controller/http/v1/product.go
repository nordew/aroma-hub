package v1

import (
	"aroma-hub/internal/application/dto"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/nordew/go-errx"

	_ "aroma-hub/internal/models"
)

func (h *Handler) initProductRoutes(api fiber.Router) {
	products := api.Group("/products")

	products.Get("/", h.listProducts)
	products.Post("/", h.createProduct)
	products.Delete("/:id", h.deleteProduct)
}

// @Summary List products
// @Description Get a list of products with optional filtering
// @Tags products
// @Accept json
// @Produce json
// @Param id query string false "Product ID"
// @Param categoryId query string false "Category ID"
// @Param categoryName query string false "Category name"
// @Param brand query string false "Brand name"
// @Param name query string false "Product name"
// @Param priceFrom query integer false "Minimum price"
// @Param priceTo query integer false "Maximum price"
// @Param stockAmountFrom query integer false "Minimum stock amount"
// @Param stockAmountTo query integer false "Maximum stock amount"
// @Success 200 {object} []models.Product "List of products"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /products [get]
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

// @Summary Create product
// @Description Add a new product to the inventory
// @Tags products
// @Accept json
// @Produce json
// @Param product body dto.CreateProductRequest true "Product information"
// @Success 201 "Created successfully"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /products [post]
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

// @Summary Delete product
// @Description Remove a product from the inventory
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 404 {object} errx.Error "Not found"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /products/{id} [delete]
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
