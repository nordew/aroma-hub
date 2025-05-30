package v1

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/consts"
	"context"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/nordew/go-errx"

	_ "aroma-hub/internal/models"
)

func (h *Handler) initProductRoutes(api fiber.Router) {
	products := api.Group("/products")

	products.Get("/", h.listProducts)
	products.Get("/brands", h.listBrands)
	products.Get("/best-sellers", h.listBestSellers)

	products.Use(h.middleware.Auth())
	products.Post("/", h.createProduct)
	products.Delete("/:id", h.deleteProduct)
	products.Patch("/:id", h.updateProduct)
	products.Patch("/:id/set-image", h.setImage)
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

// @Summary Get best sellers
// @Description Get a list of best-selling products
// @Tags products
// @Accept json
// @Produce json
// @Param limit query integer false "Limit number of results"
// @Param page query integer false "Page number for pagination"
// @Success 200 {object} []models.Product "List of best-selling products"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /products/best-sellers [get]
func (h *Handler) listBestSellers(c *fiber.Ctx) error {
	const op = "listBestSellers"

	filter := dto.ListProductFilter{
		OnlyBestSellers: true,
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
// @Accept multipart/form-data
// @Produce json
// @Param image formData file false "Product image file"
// @Param data formData string true "Product information in JSON format"
// @Success 201 {object} string "Created successfully"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /products [post]
func (h *Handler) createProduct(c *fiber.Ctx) error {
	const op = "createProduct"

	var input dto.CreateProductRequest
	err := c.BodyParser(&input)
	if err != nil {
		return handleError(c, err, op)
	}

	err = h.service.CreateProduct(context.Background(), input)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusCreated, "")
}

// @Summary List brands
// @Description Get a list of product brands
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {object} dto.BrandResponse "List of brands"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /products/brands [get]
func (h *Handler) listBrands(c *fiber.Ctx) error {
	const op = "listBrands"

	brands, err := h.service.ListBrands(context.Background())
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusOK, brands)
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

	return writeResponse(c, fiber.StatusNoContent, id)
}

// @Summary Update product
// @Description Update a product in the inventory
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body dto.UpdateProductRequest true "Product information"
// @Success 204 "No Content"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 404 {object} errx.Error "Not found"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /products/{id} [patch]
func (h *Handler) updateProduct(c *fiber.Ctx) error {
	const op = "updateProduct"

	id := c.Params("id")
	if id == "" {
		return handleError(c, errx.NewBadRequest().WithDescription("id is empty"), op)
	}

	var input dto.UpdateProductRequest
	if err := c.BodyParser(&input); err != nil {
		return handleError(c, err, op)
	}

	input.ID = id

	if input.ID == "" {
		return handleError(c, errx.NewBadRequest().WithDescription("id is empty"), op)
	}

	err := h.service.UpdateProduct(context.Background(), input)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusNoContent, "")
}

// @Summary Set product image
// @Description Set the image of a product in the inventory
// @Tags products
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Product ID"
// @Param image formData file true "Product image file"
// @Success 204 "No Content"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 404 {object} errx.Error "Not found"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /products/{id}/set-image [patch]
func (h *Handler) setImage(c *fiber.Ctx) error {
	const op = "setImage"

	productID := c.Params("id")
	if productID == "" {
		return handleError(c, errx.NewBadRequest().WithDescription("id is empty"), op)
	}

	fileHeader, err := c.FormFile(consts.ImagePrefix)
	if err != nil {
		return handleError(c, errx.NewBadRequest().WithDescription("failed to get image file: "+err.Error()), op)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return handleError(c, errx.NewBadRequest().WithDescription("failed to open image file: "+err.Error()), op)
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return handleError(c, errx.NewBadRequest().WithDescription("failed to read image file: "+err.Error()), op)
	}

	err = h.service.SetProductImage(c.Context(), productID, imageBytes)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusNoContent, nil)
}
