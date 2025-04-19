package v1

import (
	"aroma-hub/internal/application/dto"
	"context"

	_ "aroma-hub/internal/models"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) initCategoryRoutes(api fiber.Router) {
	categories := api.Group("/categories")
	categories.Get("/", h.listCategories)
	// Add other category routes here (create, delete, etc.)
}

// @Summary List categories
// @Description Get a list of categories with optional filtering
// @Tags categories
// @Accept json
// @Produce json
// @Param id query string false "Category ID"
// @Param name query string false "Category name"
// @Param limit query integer false "Limit number of results"
// @Param page query integer false "Page number for pagination"
// @Success 200 {object} []models.Category "List of categories"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /categories [get]
func (h *Handler) listCategories(c *fiber.Ctx) error {
	const op = "listCategories"
	var filter dto.ListCategoryFilter
	if err := c.QueryParser(&filter); err != nil {
		return handleError(c, err, op)
	}
	resp, err := h.service.ListCategories(context.Background(), filter)
	if err != nil {
		return handleError(c, err, op)
	}
	return writeResponse(c, fiber.StatusOK, resp)
}
