package storage

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nordew/go-errx"
	"strings"
)

type CategoryStorage struct {
	pool *pgxpool.Pool
	sb   squirrel.StatementBuilderType
}

func NewCategoryStorage(pool *pgxpool.Pool) *CategoryStorage {
	return &CategoryStorage{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (s *CategoryStorage) Create(ctx context.Context, category models.Category) (*models.Category, error) {
	query := `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`

	var result models.Category
	err := s.pool.QueryRow(ctx, query, category.Name).Scan(
		&result.ID,
		&result.Name,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errx.NewAlreadyExists().WithDescriptionAndCause(
				fmt.Sprintf("category with name '%s' already exists", category.Name),
				err,
			)
		}

		return nil, errx.NewInternal().WithDescriptionAndCause(
			"failed to create category",
			err,
		)
	}

	return &result, nil
}

func (s *CategoryStorage) List(ctx context.Context, filter dto.ListCategoryFilter) ([]models.Category, int64, error) {
	baseQuery, countQuery := s.buildSearchQuery(filter)

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

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to build count query",
			err,
		)
	}

	var totalCount int64
	err = s.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to count categories",
			err,
		)
	}

	if totalCount == 0 {
		return []models.Category{}, 0, nil
	}

	sql, args, err := baseQuery.ToSql()
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to build query",
			err,
		)
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to query categories",
			err,
		)
	}
	defer rows.Close()

	categories, err := s.scanCategories(rows)
	if err != nil {
		return nil, 0, err
	}

	return categories, totalCount, nil
}

func (s *CategoryStorage) buildSearchQuery(filter dto.ListCategoryFilter) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	baseQuery := s.sb.Select(
		"id",
		"name",
		"created_at",
		"updated_at",
	).From("categories")

	countQuery := s.sb.Select("COUNT(*)").From("categories")

	if filter.ID > 0 {
		baseQuery = baseQuery.Where(squirrel.Eq{"id": filter.ID})
		countQuery = countQuery.Where(squirrel.Eq{"id": filter.ID})
	}

	if filter.Name != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"name": "%" + filter.Name + "%"})
		countQuery = countQuery.Where(squirrel.ILike{"name": "%" + filter.Name + "%"})
	}

	return baseQuery, countQuery
}

func (s *CategoryStorage) scanCategories(rows pgx.Rows) ([]models.Category, error) {
	var categories []models.Category

	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, errx.NewInternal().WithDescriptionAndCause(
				"failed to scan category",
				err,
			)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, errx.NewInternal().WithDescriptionAndCause(
			"rows error",
			err,
		)
	}

	return categories, nil
}

func (s *CategoryStorage) Delete(ctx context.Context, id int) error {
	result, err := s.pool.Exec(ctx, "DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause(
			"category deletion failed",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return errx.NewNotFound().WithDescription("category not found")
	}

	return nil
}
