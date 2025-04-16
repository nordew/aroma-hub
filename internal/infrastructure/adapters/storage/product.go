package storage

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nordew/go-errx"
	"strings"
)

type ProductStorage struct {
	conn *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewProductStorage(conn *pgxpool.Pool) *ProductStorage {
	return &ProductStorage{
		conn: conn,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (s *ProductStorage) Create(ctx context.Context, product models.Product) error {
	_, err := s.conn.Exec(
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

func (s *ProductStorage) List(ctx context.Context, filter dto.ListProductFilter) ([]models.Product, int64, error) {
	baseQuery, countQuery := s.buildSearchQuery(filter)

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

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause("failed to build count query", err)
	}

	var totalCount int64
	err = s.conn.QueryRow(ctx, countSQL, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause("failed to get total count", err)
	}
	if totalCount == 0 {
		return []models.Product{}, 0, errx.NewNotFound().WithDescription("no products found")
	}

	sql, args, err := baseQuery.ToSql()
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause("failed to build query", err)
	}

	rows, err := s.conn.Query(ctx, sql, args...)
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

func (s *ProductStorage) buildSearchQuery(filter dto.ListProductFilter) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	baseQuery := s.sb.Select(
		"id",
		"category_id",
		"brand",
		"name",
		"image_url",
		"description",
		"composition",
		"characteristics",
		"price",
		"stock_amount",
		"created_at",
		"updated_at",
	).From("products")

	countQuery := s.sb.Select("COUNT(*)").From("products")

	if filter.ID != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"id": filter.ID})
		countQuery = countQuery.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.CategoryID != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"category_id": filter.CategoryID})
		countQuery = countQuery.Where(squirrel.Eq{"category_id": filter.CategoryID})
	}
	if filter.Brand != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"brand": "%" + filter.Brand + "%"})
		countQuery = countQuery.Where(squirrel.ILike{"brand": "%" + filter.Brand + "%"})
	}
	if filter.Name != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"name": "%" + filter.Name + "%"})
		countQuery = countQuery.Where(squirrel.ILike{"name": "%" + filter.Name + "%"})
	}
	if filter.PriceFrom > 0 {
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"price": filter.PriceFrom})
		countQuery = countQuery.Where(squirrel.GtOrEq{"price": filter.PriceFrom})
	}
	if filter.PriceTo > 0 {
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"price": filter.PriceTo})
		countQuery = countQuery.Where(squirrel.LtOrEq{"price": filter.PriceTo})
	}
	if filter.StockAmountFrom > 0 {
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"stock_amount": filter.StockAmountFrom})
		countQuery = countQuery.Where(squirrel.GtOrEq{"stock_amount": filter.StockAmountFrom})
	}
	if filter.StockAmountTo > 0 {
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"stock_amount": filter.StockAmountTo})
		countQuery = countQuery.Where(squirrel.LtOrEq{"stock_amount": filter.StockAmountTo})
	}

	return baseQuery, countQuery
}

func (s *ProductStorage) scanProducts(rows pgx.Rows) ([]models.Product, error) {
	var products []models.Product

	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID,
			&p.CategoryID,
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

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

func (s *ProductStorage) Delete(ctx context.Context, id string) error {
	_, err := s.conn.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errx.NewNotFound().WithDescription("product not found")
		}

		return errx.NewInternal().WithDescriptionAndCause("product deletion failed", err)
	}

	return nil
}
