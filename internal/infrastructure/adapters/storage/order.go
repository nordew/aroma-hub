package storage

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/nordew/go-errx"
)

func (s *Storage) CreateOrder(ctx context.Context, order models.Order) (models.Order, error) {
	query := `
		INSERT INTO orders (
			id,
			full_name,
			phone_number,
			address,
			payment_method,
			promo_code,
			contact_type,
			amount_to_pay,
			status,
			created_at,
			updated_at
		)

		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)

		RETURNING

		id,
		full_name,
		phone_number,
		address,
		payment_method,
		promo_code,
		contact_type,
		amount_to_pay,
		status,
		created_at,
		updated_at
	`
	var result models.Order
	err := s.GetQuerier().QueryRow(
		ctx, query,
		order.ID,
		order.FullName,
		order.PhoneNumber,
		order.Address,
		order.PaymentMethod,
		order.PromoCode,
		order.ContactType,
		order.AmountToPay,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(
		&result.ID,
		&result.FullName,
		&result.PhoneNumber,
		&result.Address,
		&result.PaymentMethod,
		&result.PromoCode,
		&result.ContactType,
		&result.AmountToPay,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return models.Order{}, errx.NewInternal().WithDescriptionAndCause(
			"failed to create order",
			err,
		)
	}
	return result, nil
}

func (s *Storage) ListOrders(ctx context.Context, filter dto.ListOrderFilter) ([]models.Order, int64, error) {
	baseQuery, countQuery := s.buildSearchOrderQuery(filter)
	limit := uint(10)
	if filter.Limit > 0 && filter.Limit <= 100 {
		limit = filter.Limit
	}
	offset := uint(0)
	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}
	baseQuery = baseQuery.OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	var totalCount int64
	countRow := s.squirrelHelper.QueryRow(ctx, s.GetQuerier(), countQuery)
	err := countRow.Scan(&totalCount)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to count orders",
			err,
		)
	}

	if totalCount == 0 {
		return []models.Order{}, 0, errx.NewNotFound().WithDescription("no orders found")
	}

	rows, err := s.squirrelHelper.Query(ctx, s.GetQuerier(), baseQuery)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to query orders",
			err,
		)
	}
	defer rows.Close()

	orders, err := s.scanOrders(rows)
	if err != nil {
		return nil, 0, err
	}

	return orders, totalCount, nil
}

func (s *Storage) buildSearchOrderQuery(filter dto.ListOrderFilter) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	baseQuery := s.Builder().Select(
		"id",
		"full_name",
		"phone_number",
		"address",
		"payment_method",
		"promo_code",
		"contact_type",
		"amount_to_pay",
		"status",
		"created_at",
		"updated_at",
	).From("orders")

	countQuery := s.Builder().Select("COUNT(*)").From("orders")

	if len(filter.IDs) > 0 {
		baseQuery = baseQuery.Where(squirrel.Eq{"id": filter.IDs})
		countQuery = countQuery.Where(squirrel.Eq{"id": filter.IDs})
	}
	if filter.UserID != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"user_id": filter.UserID})
		countQuery = countQuery.Where(squirrel.Eq{"user_id": filter.UserID})
	}
	if filter.PaymentMethod != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"payment_method": filter.PaymentMethod})
		countQuery = countQuery.Where(squirrel.Eq{"payment_method": filter.PaymentMethod})
	}
	if filter.ContactType != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"contact_type": filter.ContactType})
		countQuery = countQuery.Where(squirrel.Eq{"contact_type": filter.ContactType})
	}
	if filter.Status != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"status": filter.Status})
		countQuery = countQuery.Where(squirrel.Eq{"status": filter.Status})
	}
	if filter.FromDate != nil {
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"created_at": filter.FromDate})
		countQuery = countQuery.Where(squirrel.GtOrEq{"created_at": filter.FromDate})
	}
	if filter.ToDate != nil {
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"created_at": filter.ToDate})
		countQuery = countQuery.Where(squirrel.LtOrEq{"created_at": filter.ToDate})
	}

	return baseQuery, countQuery
}

func (s *Storage) scanOrders(rows pgx.Rows) ([]models.Order, error) {
	var orders []models.Order

	for rows.Next() {
		var order models.Order

		err := rows.Scan(
			&order.ID,
			&order.FullName,
			&order.PhoneNumber,
			&order.Address,
			&order.PaymentMethod,
			&order.PromoCode,
			&order.ContactType,
			&order.AmountToPay,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)

		if err != nil {
			return nil, errx.NewInternal().WithDescriptionAndCause(
				"failed to scan order",
				err,
			)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, errx.NewInternal().WithDescriptionAndCause(
			"rows error",
			err,
		)
	}

	return orders, nil
}

func (s *Storage) UpdateOrder(ctx context.Context, input dto.UpdateOrderRequest) error {
	exists, err := s.orderExists(ctx, input.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errx.NewNotFound().WithDescription("order not found")
	}

	var (
		setClauses []string
		args       []any
		paramCount = 1
	)

	if input.FullName != "" {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", paramCount))
		args = append(args, input.FullName)
		paramCount++
	}

	if input.PhoneNumber != "" {
		setClauses = append(setClauses, fmt.Sprintf("phone_number = $%d", paramCount))
		args = append(args, input.PhoneNumber)
		paramCount++
	}

	if input.Address != "" {
		setClauses = append(setClauses, fmt.Sprintf("address = $%d", paramCount))
		args = append(args, input.Address)
		paramCount++
	}

	if input.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", paramCount))
		args = append(args, input.Status)
		paramCount++
	}

	if input.PaymentMethod != "" {
		setClauses = append(setClauses, fmt.Sprintf("payment_method = $%d", paramCount))
		args = append(args, input.PaymentMethod)
		paramCount++
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE orders
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), paramCount)

	args = append(args, input.ID)

	_, err = s.GetQuerier().Exec(ctx, query, args...)
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause("order update failed", err)
	}

	return nil
}

func (s *Storage) orderExists(ctx context.Context, orderID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`

	var exists bool
	err := s.GetQuerier().QueryRow(ctx, query, orderID).Scan(&exists)
	if err != nil {
		return false, errx.NewInternal().WithDescriptionAndCause("failed to check order existence", err)
	}

	return exists, nil
}

func (s *Storage) DeleteOrder(ctx context.Context, id string) error {
	result, err := s.GetQuerier().Exec(ctx, "DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause(
			"order deletion failed",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return errx.NewNotFound().WithDescription(fmt.Sprintf("order with id '%s' not found", id))
	}

	return nil
}
