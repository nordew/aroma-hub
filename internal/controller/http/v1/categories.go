package v1

import (
	"aroma-hub/internal/application/dto"
	"context"
	"errors"

	_ "aroma-hub/internal/models"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) initCategoryRoutes(api fiber.Router) {
	categories := api.Group("/categories")

	categories.Use(h.middleware.Auth())

	categories.Get("/", h.listCategories)
	categories.Post("/", h.createCategory)
	categories.Delete("/:id", h.deleteCategory)
}

// @Summary Create category
// @Description Create a new category
// @Tags categories
// @Accept json
// @Produce json
// @Param input body dto.CreateCategoryRequest true "Category information"
// @Success 201 {object} dto.CreateCategoryRequest "Created category"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 409 {object} errx.Error "Category already exists"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /categories [post]
func (h *Handler) createCategory(c *fiber.Ctx) error {
	const op = "createCategory"

	var input dto.CreateCategoryRequest
	if err := c.BodyParser(&input); err != nil {
		return handleError(c, err, op)
	}

	if err := h.service.CreateCategory(context.Background(), input); err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusCreated, input)
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

// @Summary Delete category
// @Description Delete a category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 204 "No content"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 404 {object} errx.Error "Not found"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /categories/{id} [delete]
func (h *Handler) deleteCategory(c *fiber.Ctx) error {
	const op = "deleteCategory"

	id := c.Params("id")
	if id == "" {
		return handleError(c, errors.New("category ID is required"), op)
	}

	if err := h.service.DeleteCategory(context.Background(), id); err != nil {
		return handleError(c, err, op)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
