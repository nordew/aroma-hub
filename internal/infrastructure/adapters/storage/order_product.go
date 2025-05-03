package storage

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nordew/go-errx"
)

func (s *Storage) CreateOrderProduct(ctx context.Context, orderProduct models.OrderProduct) error {
	query := `
		INSERT INTO order_products (
			order_id,
			product_id,
			quantity,
			volume
		)
		VALUES ($1, $2, $3, $4)
	`
	_, err := s.GetQuerier().Exec(ctx, query,
		orderProduct.OrderID,
		orderProduct.ProductID,
		orderProduct.Quantity,
		orderProduct.Volume,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == uniqueViolationCode {
				if strings.Contains(pgErr.ConstraintName, "order_products_pkey") {
					return errx.NewAlreadyExists().WithDescriptionAndCause(
						fmt.Sprintf("order product with order_id '%s' and product_id '%s' already exists",
							orderProduct.OrderID, orderProduct.ProductID),
						err,
					)
				}

				return errx.NewAlreadyExists().WithDescriptionAndCause(
					"order product with these details already exists",
					err,
				)
			}
		}

		return errx.NewInternal().WithDescriptionAndCause(
			"failed to create order product",
			err,
		)
	}
	return nil
}

func (s *Storage) ListOrderProducts(ctx context.Context, filter dto.ListOrderProductFilter) ([]models.OrderProduct, int64, error) {
	baseQuery, countQuery := s.buildSearchOrderProductQuery(filter)

	limit := uint(10)
	if filter.Limit > 0 && filter.Limit <= 100 {
		limit = filter.Limit
	}

	offset := uint(0)
	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}

	baseQuery = baseQuery.Limit(uint64(limit)).
		Offset(uint64(offset))

	var totalCount int64
	countRow := s.squirrelHelper.QueryRow(ctx, s.GetQuerier(), countQuery)
	err := countRow.Scan(&totalCount)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to count order products",
			err,
		)
	}

	if totalCount == 0 {
		return []models.OrderProduct{}, 0, errx.NewNotFound().WithDescription("no order products found")
	}

	rows, err := s.squirrelHelper.Query(ctx, s.GetQuerier(), baseQuery)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to query order products",
			err,
		)
	}
	defer rows.Close()

	orderProducts, err := s.scanOrderProducts(rows)
	if err != nil {
		return nil, 0, err
	}

	return orderProducts, totalCount, nil
}

func (s *Storage) buildSearchOrderProductQuery(filter dto.ListOrderProductFilter) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	baseQuery := s.Builder().Select(
		"order_id",
		"product_id",
		"quantity",
		"volume",
	).From("order_products")

	countQuery := s.Builder().Select("COUNT(*)").From("order_products")

	if len(filter.OrderIDs) > 0 {
		baseQuery = baseQuery.Where(squirrel.Eq{"order_id": filter.OrderIDs})
		countQuery = countQuery.Where(squirrel.Eq{"order_id": filter.OrderIDs})
	}

	if len(filter.ProductIDs) > 0 {
		baseQuery = baseQuery.Where(squirrel.Eq{"product_id": filter.ProductIDs})
		countQuery = countQuery.Where(squirrel.Eq{"product_id": filter.ProductIDs})
	}

	return baseQuery, countQuery
}

func (s *Storage) scanOrderProducts(rows pgx.Rows) ([]models.OrderProduct, error) {
	var orderProducts []models.OrderProduct

	for rows.Next() {
		var orderProduct models.OrderProduct

		err := rows.Scan(
			&orderProduct.OrderID,
			&orderProduct.ProductID,
			&orderProduct.Quantity,
			&orderProduct.Volume,
		)

		if err != nil {
			return nil, errx.NewInternal().WithDescriptionAndCause(
				"failed to scan order product",
				err,
			)
		}

		orderProducts = append(orderProducts, orderProduct)
	}

	if err := rows.Err(); err != nil {
		return nil, errx.NewInternal().WithDescriptionAndCause(
			"rows error",
			err,
		)
	}

	return orderProducts, nil
}

func (s *Storage) DeleteOrderProduct(ctx context.Context, orderID, productID string) error {
	result, err := s.GetQuerier().Exec(
		ctx,
		"DELETE FROM order_products WHERE order_id = $1 AND product_id = $2",
		orderID,
		productID,
	)
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause(
			"order product deletion failed",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return errx.NewNotFound().WithDescription(
			fmt.Sprintf("order product with order_id '%s' and product_id '%s' not found", orderID, productID),
		)
	}

	return nil
}
