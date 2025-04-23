package v1

import (
	"aroma-hub/internal/application/dto"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/nordew/go-errx"
)

func (h *Handler) initPromocodeRoutes(api fiber.Router) {
	promocodes := api.Group("/promocodes")
	promocodes.Post("/", h.createPromocode)
	promocodes.Get("/", h.listPromocodes)
	promocodes.Delete("/:id", h.deletePromocode)
}

// @Summary Create promocode
// @Description Create a new promocode with discount
// @Tags promocodes
// @Accept json
// @Produce json
// @Param promocode body dto.CreatePromocodeRequest true "Promocode information"
// @Success 201 "Created successfully"
// @Failure 400 {object} errx.Error "Validation error"
// @Failure 409 {object} errx.Error "Promocode already exists"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /promocodes [post]
func (h *Handler) createPromocode(c *fiber.Ctx) error {
	const op = "createPromocode"
	var input dto.CreatePromocodeRequest
	if err := c.BodyParser(&input); err != nil {
		return handleError(c, err, op)
	}
	if err := h.service.CreatePromocode(context.Background(), input); err != nil {
		return handleError(c, err, op)
	}
	return writeResponse(c, fiber.StatusCreated, nil)
}

// @Summary List promocodes
// @Description Get a list of promocodes with optional filtering
// @Tags promocodes
// @Accept json
// @Produce json
// @Param id query string false "Promocode ID"
// @Param code query string false "Promocode code"
// @Param discountFrom query integer false "Minimum discount percentage"
// @Param discountTo query integer false "Maximum discount percentage"
// @Param active query boolean false "Filter for active promocodes (not expired)"
// @Param limit query integer false "Number of items per page (default: 10, max: 100)"
// @Param page query integer false "Page number (default: 1)"
// @Success 200 {object} dto.ListPromocodesResponse "List of promocodes"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 404 {object} errx.Error "No promocodes found"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /promocodes [get]
func (h *Handler) listPromocodes(c *fiber.Ctx) error {
	const op = "listPromocodes"
	var input dto.ListPromocodeFilter
	if err := c.QueryParser(&input); err != nil {
		return handleError(c, err, op)
	}

	resp, err := h.service.ListPromocodes(context.Background(), input)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusOK, resp)
}

// @Summary Delete promocode
// @Description Delete a promocode by its ID
// @Tags promocodes
// @Accept json
// @Produce json
// @Param id path string true "Promocode ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 404 {object} errx.Error "Promocode not found"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /promocodes/{id} [delete]
func (h *Handler) deletePromocode(c *fiber.Ctx) error {
	const op = "deletePromocode"
	id := c.Params("id")
	if id == "" {
		return handleError(c, errx.NewBadRequest().WithDescription("id is empty"), op)
	}

	if err := h.service.DeletePromocode(context.Background(), id); err != nil {
		return handleError(c, err, op)
	}
	return writeResponse(c, fiber.StatusNoContent, nil)
}
