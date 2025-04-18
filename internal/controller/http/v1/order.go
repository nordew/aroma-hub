package v1

import (
	"aroma-hub/internal/application/dto"
	"context"
	"time"

	_ "aroma-hub/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/nordew/go-errx"
)

func (h *Handler) initOrderRoutes(api fiber.Router) {
	orders := api.Group("/orders")
	orders.Get("/", h.listOrders)
	orders.Post("/", h.createOrder)
	orders.Delete("/:id", h.deleteOrder)
}

// @Summary List orders
// @Description Get a list of orders with optional filtering
// @Tags orders
// @Accept json
// @Produce json
// @Param id query string false "Order ID"
// @Param userId query string false "User ID"
// @Param paymentMethod query string false "Payment method (IBAN, сash_on_delivery)"
// @Param contactType query string false "Contact type (telegram, phone)"
// @Param status query string false "Order status (pending, processing, completed, cancelled)"
// @Param fromDate query string false "Start date for filtering (format: YYYY-MM-DD)"
// @Param toDate query string false "End date for filtering (format: YYYY-MM-DD)"
// @Param limit query integer false "Number of items per page (default: 10, max: 100)"
// @Param page query integer false "Page number (default: 1)"
// @Success 200 {object} dto.ListOrdersResponse "List of orders"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 404 {object} errx.Error "No orders found"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /orders [get]
func (h *Handler) listOrders(c *fiber.Ctx) error {
	const op = "listOrders"

	var filter dto.ListOrderFilter
	if err := c.QueryParser(&filter); err != nil {
		return handleError(c, err, op)
	}

	if fromDateStr := c.Query("fromDate"); fromDateStr != "" {
		fromDate, err := time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			return handleError(c, errx.NewBadRequest().WithDescription("invalid fromDate format"), op)
		}

		filter.FromDate = &fromDate
	}

	if toDateStr := c.Query("toDate"); toDateStr != "" {
		toDate, err := time.Parse("2006-01-02", toDateStr)
		if err != nil {
			return handleError(c, errx.NewBadRequest().WithDescription("invalid toDate format"), op)
		}

		toDate = toDate.Add(24*time.Hour - time.Second)
		filter.ToDate = &toDate
	}

	resp, err := h.service.ListOrders(context.Background(), filter)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusOK, resp)
}

// @Summary Create order
// @Description Create a new order
// @Tags orders
// @Accept json
// @Produce json
// @Param order body dto.CreateOrderRequest true "Order information"
// @Success 201 "Created"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /orders [post]
func (h *Handler) createOrder(c *fiber.Ctx) error {
	const op = "createOrder"

	var input dto.CreateOrderRequest
	if err := c.BodyParser(&input); err != nil {
		return handleError(c, errx.NewBadRequest().WithDescriptionAndCause("invalid request body", err), op)
	}

	if err := h.service.CreateOrder(context.Background(), input); err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusCreated, "")
}

// @Summary Delete order
// @Description Delete an order by ID
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 204 "No Content"
// @Failure 400 {object} errx.Error "Bad request"
// @Failure 404 {object} errx.Error "Order not found"
// @Failure 500 {object} errx.Error "Internal server error"
// @Router /orders/{id} [delete]
func (h *Handler) deleteOrder(c *fiber.Ctx) error {
	const op = "deleteOrder"

	id := c.Params("id")
	if id == "" {
		return handleError(c, errx.NewBadRequest().WithDescription("id is empty"), op)
	}

	err := h.service.DeleteOrder(context.Background(), id)
	if err != nil {
		return handleError(c, err, op)
	}

	return writeResponse(c, fiber.StatusNoContent, nil)
}
