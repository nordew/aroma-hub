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
	ErrCreateOrder                    = "Failed to create order"
	ErrGetProduct                     = "Failed to get product"
	ErrProductNotFound                = "Product not found"
	ErrInsufficientStock              = "Insufficient stock"
	ErrOrderValidation                = "Order validation failed"
	ErrUpdateStock                    = "Failed to update stock"
	ErrPersistOrder                   = "Failed to save order"
	ErrPromoCodeExpired               = "Promo code has expired"
	ErrOnlyPendingOrdersCanBeCanceled = "only pending orders can be canceled"
)

type OrderData struct {
	TotalAmount   decimal.Decimal
	StockUpdates  map[string]uint
	OrderProducts []models.OrderProduct
}

func (s *Service) CreateOrder(ctx context.Context, input dto.CreateOrderRequest) error {
	if err := s.validateOrderInput(ctx, input); err != nil {
		return err
	}

	productInfo, err := s.prepareProductInfo(ctx, input.ProductItems)
	if err != nil {
		return err
	}

	orderID := uuid.New().String()

	orderData, err := s.calculateOrderData(input.ProductItems, orderID, productInfo.productByID)
	if err != nil {
		return err
	}

	order, err := models.NewOrder(
		orderID,
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

	if err := s.executeOrderTransaction(ctx, order, orderData); err != nil {
		return err
	}

	msgText := fmt.Sprintf("Order %s placed", orderID)
	go s.messagingProvider.BroadcastMessage(ctx, msgText)

	return nil
}
func (s *Service) validateOrderInput(ctx context.Context, input dto.CreateOrderRequest) error {
	if len(input.ProductItems) == 0 {
		return errx.NewBadRequest().WithDescription("order must contain at least one item")
	}

	if input.PromoCode != "" {
		if err := s.validatePromoCode(ctx, input.PromoCode); err != nil {
			return err
		}
	}

	return nil
}

type productInfoResult struct {
	productIDs   []string
	productByID  map[string]models.Product
	quantityByID map[string]uint
}

func (s *Service) prepareProductInfo(ctx context.Context, productItems []dto.ProductOrderItem) (productInfoResult, error) {
	result := productInfoResult{
		productIDs:   make([]string, 0, len(productItems)),
		productByID:  make(map[string]models.Product),
		quantityByID: make(map[string]uint, len(productItems)),
	}

	productIDSet := make(map[string]struct{}, len(productItems))
	for _, item := range productItems {
		result.quantityByID[item.ID] = item.Quantity

		if _, exists := productIDSet[item.ID]; !exists {
			productIDSet[item.ID] = struct{}{}
			result.productIDs = append(result.productIDs, item.ID)
		}
	}

	filter := dto.ListProductFilter{
		IDs:   result.productIDs,
		Limit: uint(len(result.productIDs)),
	}
	products, _, err := s.storage.ListProducts(ctx, filter)
	if err != nil {
		return productInfoResult{}, err
	}

	if len(products) != len(result.productIDs) {
		return productInfoResult{}, errx.NewNotFound().WithDescription(ErrProductNotFound)
	}

	for _, product := range products {
		result.productByID[product.ID] = product
	}

	return result, nil
}

func (s *Service) calculateOrderData(
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

		if err := s.validateProductStock(product, productItem.Quantity); err != nil {
			return OrderData{}, err
		}

		orderProduct, err := models.NewOrderProduct(
			orderID,
			product.ID,
			productItem.Quantity,
			productItem.Volume)
		if err != nil {
			return OrderData{}, fmt.Errorf("creating order product: %w", err)
		}

		result.OrderProducts = append(result.OrderProducts, orderProduct)

		itemAmount := product.Price.Mul(decimal.NewFromInt(int64(productItem.Quantity)))
		result.TotalAmount = result.TotalAmount.Add(itemAmount)

		result.StockUpdates[product.ID] = product.StockAmount - productItem.Quantity
	}

	return result, nil
}

func (s *Service) validateProductStock(product models.Product, requestedQuantity uint) error {
	if product.StockAmount == 0 {
		return errx.NewBadRequest().WithDescription(
			fmt.Sprintf("invalid quantity for product %s: must be greater than zero", product.Name))
	}

	if product.StockAmount < requestedQuantity {
		return errx.NewBadRequest().WithDescription(
			fmt.Sprintf("insufficient stock for product %s: requested %d, available %d",
				product.Name, requestedQuantity, product.StockAmount))
	}

	return nil
}

func (s *Service) executeOrderTransaction(ctx context.Context, order models.Order, orderData OrderData) error {
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

func (s *Service) UpdateOrder(ctx context.Context, input dto.UpdateOrderRequest) error {
	return s.storage.UpdateOrder(ctx, input)
}

func (s *Service) CancelOrder(ctx context.Context, id string) error {
	if err := s.transactor.ExecuteInTx(ctx, []pgxtransactor.Storage{s.storage}, func() error {
		orders, _, err := s.storage.ListOrders(ctx, dto.ListOrderFilter{
			IDs: []string{id},
		})
		if err != nil {
			return err
		}
		order := orders[0]

		if order.Status != models.OrderStatusPending {
			return errx.NewBadRequest().WithDescription(ErrOnlyPendingOrdersCanBeCanceled)
		}

		if err := s.storage.UpdateOrder(ctx, dto.UpdateOrderRequest{
			ID:     id,
			Status: models.OrderStatusCancelled,
		}); err != nil {
			return fmt.Errorf("updating order status: %w", err)
		}

		if err := s.restoreProductQuantities(ctx, id); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) restoreProductQuantities(ctx context.Context, orderID string) error {
	orderProducts, _, err := s.storage.ListOrderProducts(ctx, dto.ListOrderProductFilter{
		OrderIDs: []string{orderID},
	})
	if err != nil {
		return err
	}

	productIDs := make([]string, 0, len(orderProducts))
	for i := range orderProducts {
		productIDs = append(productIDs, orderProducts[i].ProductID)
	}

	products, _, err := s.storage.ListProducts(ctx, dto.ListProductFilter{
		IDs: productIDs,
	})
	if err != nil {
		return err
	}

	productMap := make(map[string]models.Product, len(products))
	for i := range products {
		productMap[products[i].ID] = products[i]
	}

	for _, op := range orderProducts {
		product, exists := productMap[op.ProductID]
		if !exists {
			continue
		}

		product.StockAmount += op.Quantity

		if err := s.storage.UpdateProduct(ctx, dto.UpdateProductRequest{
			ID:          product.ID,
			StockAmount: product.StockAmount,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) DeleteOrder(ctx context.Context, id string) error {
	orders, _, err := s.storage.ListOrders(ctx, dto.ListOrderFilter{
		IDs: []string{id},
	})
	if err != nil {
		return err
	}
	order := orders[0]

	switch order.Status {
	case models.OrderStatusPending:
		if err := s.CancelOrder(ctx, id); err != nil {
			return err
		}
	case models.OrderStatusCancelled:
		return errx.NewBadRequest().WithDescription("order already cancelled")
	case models.OrderStatusCompleted:
		return errx.NewBadRequest().WithDescription("order already completed")
	}

	err = s.storage.DeleteOrder(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
