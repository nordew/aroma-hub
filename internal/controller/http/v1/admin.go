package v1

import (
	"aroma-hub/internal/application/dto"
	"context"

	_ "aroma-hub/internal/models"
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) initAdminRoutes(api fiber.Router) {
	admin := api.Group("/admin")

	admin.Post("/login", h.adminLogin)
	admin.Get("/refresh", h.adminRefresh)
	admin.Get("/products", h.adminListProducts)
}

// @Summary Admin login
// @Description Admin login with OTP code
// @Tags admin
// @Accept json
// @Produce json
// @Param input body dto.AdminLoginRequest true "Admin login information"
// @Success 200 {object} dto.AdminLoginResponse "Admin login response"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 401 {object} errx.Error "Unauthorized"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /admin/login [post]
func (h *Handler) adminLogin(c *fiber.Ctx) error {
	const op = "adminLogin"

	var input dto.AdminLoginRequest
	if err := c.BodyParser(&input); err != nil {
		return handleError(c, err, op)
	}

	resp, err := h.service.AdminLogin(context.Background(), input)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusOK, resp)
}

// @Summary Admin refresh token
// @Description Refresh admin access token
// @Tags admin
// @Accept json
// @Produce json
// @Param input body dto.AdminRefreshTokenRequest true "Admin refresh token information"
// @Success 200 {object} dto.AdminRefreshTokenResponse "Admin refresh token response"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 401 {object} errx.Error "Unauthorized"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /admin/refresh [get]
func (h *Handler) adminRefresh(c *fiber.Ctx) error {
	const op = "adminRefresh"

	var input dto.AdminRefreshTokenRequest
	if err := c.QueryParser(&input); err != nil {
		return handleError(c, err, op)
	}

	resp, err := h.service.AdminRefresh(context.Background(), input)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusOK, resp)
}

// @Summary List products
// @Description Get a list of products with optional filtering (invisible included)
// @Tags admin
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
// @Router /admin/products [get]
func (h *Handler) adminListProducts(c *fiber.Ctx) error {
	const op = "listProducts"

	var filter dto.ListProductFilter
	if err := c.QueryParser(&filter); err != nil {
		return handleError(c, err, op)
	}

	filter.ShowInvisible = true

	resp, err := h.service.ListProducts(context.Background(), filter)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusOK, resp)
}
