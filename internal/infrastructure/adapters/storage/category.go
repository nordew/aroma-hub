package storage

import (
	"context"
	"fmt"
	"strings"

	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/nordew/go-errx"
)

var (
	ErrCategoryNotFound        = "category not found"
	ErrNoCategoriesFound       = "no categories found"
	ErrFailedToCreateCategory  = "failed to create category"
	ErrFailedToCountCategories = "failed to count categories"
	ErrFailedToQueryCategories = "failed to query categories"
	ErrFailedToScanCategory    = "failed to scan category"
	ErrRowsError               = "rows error"
	ErrCategoryDeletionFailed  = "category deletion failed"
)

func (s *Storage) CreateCategory(ctx context.Context, category models.Category) error {
	query := `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id
	`
	// TODO: Add id reutrn
	var id int64
	err := s.GetQuerier().QueryRow(ctx, query, category.Name).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errx.NewAlreadyExists().WithDescriptionAndCause(
				fmt.Sprintf("category with name '%s' already exists", category.Name),
				err,
			)
		}
		return errx.NewInternal().WithDescriptionAndCause(
			ErrFailedToCreateCategory,
			err,
		)
	}

	return nil
}

func (s *Storage) ListCategories(ctx context.Context, filter dto.ListCategoryFilter) ([]models.Category, int64, error) {
	baseQuery, countQuery := s.buildSearchCategoryQuery(filter)

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
	err := s.squirrelHelper.QueryRow(ctx, s.GetQuerier(), countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			ErrFailedToCountCategories,
			err,
		)
	}

	if totalCount == 0 {
		return []models.Category{}, 0, errx.NewNotFound().WithDescription(ErrNoCategoriesFound)
	}

	rows, err := s.squirrelHelper.Query(ctx, s.GetQuerier(), baseQuery)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			ErrFailedToQueryCategories,
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

func (s *Storage) buildSearchCategoryQuery(filter dto.ListCategoryFilter) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	baseQuery := s.Builder().Select(
		"id",
		"name",
		"created_at",
		"updated_at",
	).From("categories")

	countQuery := s.Builder().Select("COUNT(*)").From("categories")

	if filter.ID != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"id": filter.ID})
		countQuery = countQuery.Where(squirrel.Eq{"id": filter.ID})
	}

	if filter.Name != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"name": "%" + filter.Name + "%"})
		countQuery = countQuery.Where(squirrel.ILike{"name": "%" + filter.Name + "%"})
	}

	return baseQuery, countQuery
}

func (s *Storage) scanCategories(rows pgx.Rows) ([]models.Category, error) {
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
				ErrFailedToScanCategory,
				err,
			)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, errx.NewInternal().WithDescriptionAndCause(
			ErrRowsError,
			err,
		)
	}

	return categories, nil
}

func (s *Storage) DeleteCategory(ctx context.Context, id string) error {
	result, err := s.GetQuerier().Exec(ctx, "DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause(
			ErrCategoryDeletionFailed,
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return errx.NewNotFound().WithDescription(ErrCategoryNotFound)
	}

	return nil
}
