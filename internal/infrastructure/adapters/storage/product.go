package storage

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/nordew/go-errx"
)

func (s *Storage) CreateProduct(ctx context.Context, product models.Product) error {
	_, err := s.GetQuerier().Exec(
		ctx,
		`
		INSERT INTO products (id, category_id, brand, name, image_url, description, composition, characteristics, price, stock_amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		product.ID,
		product.CategoryID,
		product.Brand,
		product.Name,
		product.ImageURL,
		product.Description,
		product.Composition,
		product.Characteristics,
		product.Price,
		product.StockAmount,
	)
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause("product creation failed", err)
	}

	return nil
}

func (s *Storage) ListProducts(ctx context.Context, filter dto.ListProductFilter) ([]models.Product, int64, error) {
	baseQuery, countQuery := s.buildProductSearchQuery(filter)

	sortBy := "created_at"
	sortOrder := "DESC"

	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	if filter.SortOrder != "" {
		sortOrder = strings.ToUpper(filter.SortOrder)
	}

	baseQuery = baseQuery.OrderBy(fmt.Sprintf("%s %s", sortBy, sortOrder))

	limit := uint(10)
	if filter.Limit > 0 && filter.Limit <= 100 {
		limit = filter.Limit
	}

	offset := uint(0)
	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}

	baseQuery = baseQuery.Limit(uint64(limit)).Offset(uint64(offset))

	var totalCount int64
	countRow := s.squirrelHelper.QueryRow(ctx, s.GetQuerier(), countQuery)
	err := countRow.Scan(&totalCount)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause("failed to get total count", err)
	}
	if totalCount == 0 {
		return []models.Product{}, 0, errx.NewNotFound().WithDescription("no products found")
	}

	rows, err := s.squirrelHelper.Query(ctx, s.GetQuerier(), baseQuery)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause("failed to execute query", err)
	}
	defer rows.Close()

	products, err := s.scanProducts(rows)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause("failed to scan products", err)
	}

	return products, totalCount, nil
}

func (s *Storage) buildProductSearchQuery(filter dto.ListProductFilter) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	baseQuery := s.Builder().Select(
		"p.id",
		"p.category_id",
		"c.name AS category_name",
		"p.brand",
		"p.name",
		"p.image_url",
		"p.description",
		"p.composition",
		"p.characteristics",
		"p.price",
		"p.stock_amount",
		"p.created_at",
		"p.updated_at",
	).
		From("products p").
		LeftJoin("categories c ON p.category_id = c.id")

	countQuery := s.Builder().Select("COUNT(*)").From("products p").
		LeftJoin("categories c ON p.category_id = c.id")

	if len(filter.IDs) > 0 {
		baseQuery = baseQuery.Where(squirrel.Eq{"p.id": filter.IDs})
		countQuery = countQuery.Where(squirrel.Eq{"p.id": filter.IDs})
	}
	if filter.CategoryID != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"p.category_id": filter.CategoryID})
		countQuery = countQuery.Where(squirrel.Eq{"p.category_id": filter.CategoryID})
	}
	if filter.CategoryName != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"c.name": "%" + filter.CategoryName + "%"})
		countQuery = countQuery.Where(squirrel.ILike{"c.name": "%" + filter.CategoryName + "%"})
	}
	if filter.Brand != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"p.brand": "%" + filter.Brand + "%"})
		countQuery = countQuery.Where(squirrel.ILike{"p.brand": "%" + filter.Brand + "%"})
	}
	if filter.Name != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"p.name": "%" + filter.Name + "%"})
		countQuery = countQuery.Where(squirrel.ILike{"p.name": "%" + filter.Name + "%"})
	}
	if filter.PriceFrom > 0 {
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"p.price": filter.PriceFrom})
		countQuery = countQuery.Where(squirrel.GtOrEq{"p.price": filter.PriceFrom})
	}
	if filter.PriceTo > 0 {
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"p.price": filter.PriceTo})
		countQuery = countQuery.Where(squirrel.LtOrEq{"p.price": filter.PriceTo})
	}
	if filter.StockAmountFrom > 0 {
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"p.stock_amount": filter.StockAmountFrom})
		countQuery = countQuery.Where(squirrel.GtOrEq{"p.stock_amount": filter.StockAmountFrom})
	}
	if filter.StockAmountTo > 0 {
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"p.stock_amount": filter.StockAmountTo})
		countQuery = countQuery.Where(squirrel.LtOrEq{"p.stock_amount": filter.StockAmountTo})
	}

	return baseQuery, countQuery
}

func (s *Storage) scanProducts(rows pgx.Rows) ([]models.Product, error) {
	var products []models.Product

	for rows.Next() {
		var (
			p            models.Product
			categoryName sql.NullString
		)

		err := rows.Scan(
			&p.ID,
			&p.CategoryID,
			&categoryName,
			&p.Brand,
			&p.Name,
			&p.ImageURL,
			&p.Description,
			&p.Composition,
			&p.Characteristics,
			&p.Price,
			&p.StockAmount,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		if categoryName.Valid {
			p.CategoryName = categoryName.String
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

func (s *Storage) DeleteProduct(ctx context.Context, id string) error {
	result, err := s.GetQuerier().Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause("product deletion failed", err)
	}

	if result.RowsAffected() == 0 {
		return errx.NewNotFound().WithDescription("product not found")
	}

	return nil
}
