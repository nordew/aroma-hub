package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nordew/go-errx"
	pgxtransactor "github.com/nordew/pgx-transactor"
	"github.com/shopspring/decimal"
)

var (
	ErrCreateOrder       = "Failed to create order"
	ErrGetProduct        = "Failed to get product"
	ErrProductNotFound   = "Product not found"
	ErrInsufficientStock = "Insufficient stock"
	ErrOrderValidation   = "Order validation failed"
	ErrUpdateStock       = "Failed to update stock"
	ErrPersistOrder      = "Failed to save order"
	ErrPromoCodeExpired  = "Promo code has expired"
)

type OrderData struct {
	TotalAmount   decimal.Decimal
	StockUpdates  map[string]uint // Map of product ID to new stock amount
	OrderProducts []models.OrderProduct
}

// CreateOrder processes an order request through validation, stock checking,
// price calculation, and persistence in a database transaction.
//
// Algorithm:
// 1. Validate order input (items presence, promo code)
// 2. Fetch all required products in a single query
// 3. Process each item:
//   - Verify product exists
//   - Check stock availability
//   - Calculate prices
//   - Prepare stock updates
//
// 4. Create order entity
// 5. Execute within a transaction:
//   - Create order record
//   - Update product stocks
//   - Create order items
func (s *Service) CreateOrder(ctx context.Context, input dto.CreateOrderRequest) error {
	if len(input.ProductItems) == 0 {
		return errx.NewBadRequest().WithDescription("order must contain at least one item")
	}

	if input.PromoCode != "" {
		if err := s.validatePromoCode(ctx, input.PromoCode); err != nil {
			return err
		}
	}

	productIDs := make([]string, 0, len(input.ProductItems))
	productIDSet := make(map[string]struct{}, len(input.ProductItems))
	quantityByProductID := make(map[string]uint, len(input.ProductItems))

	for _, item := range input.ProductItems {
		quantityByProductID[item.ID] = item.Quantity

		if _, exists := productIDSet[item.ID]; !exists {
			productIDSet[item.ID] = struct{}{}
			productIDs = append(productIDs, item.ID)
		}
	}

	filter := dto.ListProductFilter{
		IDs:   productIDs,
		Limit: uint(len(productIDs)),
	}
	products, _, err := s.storage.ListProducts(ctx, filter)
	if err != nil {
		return err
	}

	if len(products) != len(productIDs) {
		return errx.NewNotFound().WithDescription("one or more requested products were not found")
	}

	productByID := make(map[string]models.Product, len(products))
	for _, product := range products {
		productByID[product.ID] = product
	}

	orderID := uuid.New().String()

	orderData, err := s.prepareOrderData(input.ProductItems, orderID, productByID)
	if err != nil {
		return err
	}

	order, err := models.NewOrder(
		orderID,
		input.UserID,
		input.FullName,
		input.PhoneNumber,
		input.Address,
		input.PaymentMethod,
		input.PromoCode,
		input.ContactType,
		orderData.TotalAmount,
	)
	if err != nil {
		return err
	}

	return s.transactor.ExecuteInTx(ctx, []pgxtransactor.Storage{s.storage}, func() error {
		if _, err := s.storage.CreateOrder(ctx, order); err != nil {
			return err
		}

		for productID, newStock := range orderData.StockUpdates {
			if err := s.storage.UpdateProduct(ctx, dto.UpdateProductRequest{
				ID:          productID,
				StockAmount: newStock,
			}); err != nil {
				return err
			}
		}

		for _, product := range orderData.OrderProducts {
			if err := s.storage.CreateOrderProduct(ctx, product); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) prepareOrderData(
	productItems []dto.ProductOrderItem,
	orderID string,
	productByID map[string]models.Product,
) (OrderData, error) {
	result := OrderData{
		TotalAmount:   decimal.Zero,
		StockUpdates:  make(map[string]uint, len(productByID)),
		OrderProducts: make([]models.OrderProduct, 0, len(productItems)),
	}

	for _, productItem := range productItems {
		product, exists := productByID[productItem.ID]
		if !exists {
			return OrderData{}, errx.NewNotFound().WithDescription(
				fmt.Sprintf("product %s not found", productItem.ID))
		}

		if product.StockAmount == 0 {
			return OrderData{}, errx.NewBadRequest().WithDescription(
				fmt.Sprintf("invalid quantity for product %s: must be greater than zero", product.Name))
		}

		if product.StockAmount < productItem.Quantity {
			return OrderData{}, errx.NewBadRequest().WithDescription(
				fmt.Sprintf("insufficient stock for product %s: requested %d, available %d",
					product.Name, productItem.Quantity, product.StockAmount))
		}

		orderProduct, err := models.NewOrderProduct(orderID, product.ID, productItem.Quantity)
		if err != nil {
			return OrderData{}, err
		}

		result.OrderProducts = append(result.OrderProducts, orderProduct)

		itemAmount := product.Price.Mul(decimal.NewFromInt(int64(productItem.Quantity)))
		result.TotalAmount = result.TotalAmount.Add(itemAmount)

		result.StockUpdates[product.ID] = product.StockAmount - productItem.Quantity
	}

	return result, nil
}

func (s *Service) validatePromoCode(ctx context.Context, promoCode string) error {
	promoCodes, _, err := s.storage.ListPromocodes(ctx, dto.ListPromocodeFilter{
		Code: promoCode,
	})
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause("failed to fetch promo code", err)
	}
	code := promoCodes[0]

	if code.ExpiresAt.Before(time.Now()) {
		return errx.NewForbidden().WithDescription(ErrPromoCodeExpired)
	}

	return nil
}

func (s *Service) ListOrders(ctx context.Context, filter dto.ListOrderFilter) (dto.ListOrdersResponse, error) {
	orders, total, err := s.storage.ListOrders(ctx, filter)
	if err != nil {
		return dto.ListOrdersResponse{}, fmt.Errorf("failed to list orders: %w", err)
	}

	if len(orders) == 0 {
		return dto.ListOrdersResponse{Orders: []models.Order{}, Total: total}, nil
	}

	if err := s.enrichOrdersWithProducts(ctx, orders); err != nil {
		return dto.ListOrdersResponse{}, err
	}

	return dto.ListOrdersResponse{
		Orders: orders,
		Total:  total,
	}, nil
}

func (s *Service) enrichOrdersWithProducts(ctx context.Context, orders []models.Order) error {
	orderIDs := make([]string, len(orders))
	for i := range orders {
		order := &orders[i]
		orderIDs[i] = order.ID
	}

	allOrderProducts, _, err := s.storage.ListOrderProducts(ctx, dto.ListOrderProductFilter{
		OrderIDs: orderIDs,
	})
	if err != nil {
		if errx.IsCode(err, errx.NotFound) {
			for i := range orders {
				order := &orders[i]
				order.Products = []models.Product{}
			}

			return nil
		}

		return fmt.Errorf("failed to fetch order products: %w", err)
	}

	orderToProductIDs := make(map[string][]string)
	allProductIDs := make([]string, 0, len(allOrderProducts))

	for _, op := range allOrderProducts {
		orderToProductIDs[op.OrderID] = append(orderToProductIDs[op.OrderID], op.ProductID)
		allProductIDs = append(allProductIDs, op.ProductID)
	}

	if len(allProductIDs) == 0 {
		for i := range orders {
			order := &orders[i]
			order.Products = []models.Product{}
		}

		return nil
	}

	products, _, err := s.storage.ListProducts(ctx, dto.ListProductFilter{
		IDs:   allProductIDs,
		Limit: uint(len(allProductIDs)),
	})
	if err != nil {
		if errx.IsCode(err, errx.NotFound) {
			for i := range orders {
				order := &orders[i]
				order.Products = []models.Product{}
			}

			return nil
		}

		return fmt.Errorf("failed to fetch products: %w", err)
	}

	productMap := make(map[string]models.Product, len(products))
	for _, product := range products {
		productMap[product.ID] = product
	}

	for i := range orders {
		order := &orders[i]
		productIDs, exists := orderToProductIDs[order.ID]
		if !exists {
			order.Products = []models.Product{}
			continue
		}

		orderProducts := make([]models.Product, 0, len(productIDs))

		for _, productID := range productIDs {
			if product, ok := productMap[productID]; ok {
				orderProducts = append(orderProducts, product)
			}
		}

		order.Products = orderProducts
	}

	return nil
}

func (s *Service) DeleteOrder(ctx context.Context, id string) error {
	err := s.storage.DeleteOrder(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
