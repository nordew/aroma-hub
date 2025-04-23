package storage

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nordew/go-errx"
)

func (s *Storage) CreatePromocode(ctx context.Context, promocode models.Promocode) error {
	query := `
		INSERT INTO promocodes (
			id,
			code,
			discount,
			expires_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.GetQuerier().Exec(ctx, query,
		promocode.ID,
		promocode.Code,
		promocode.Discount,
		promocode.ExpiresAt,
		promocode.CreatedAt,
		promocode.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == uniqueViolationCode {
				if strings.Contains(pgErr.ConstraintName, "promocodes_code_key") {
					return errx.NewAlreadyExists().WithDescriptionAndCause(
						fmt.Sprintf("promocode with code '%s' already exists", promocode.Code),
						err,
					)
				}

				return errx.NewAlreadyExists().WithDescriptionAndCause(
					"promocode with these details already exists",
					err,
				)
			}
		}

		return errx.NewInternal().WithDescriptionAndCause(
			"failed to create promocode",
			err,
		)
	}
	return nil
}

func (s *Storage) ListPromocodes(ctx context.Context, filter dto.ListPromocodeFilter) ([]models.Promocode, int64, error) {
	baseQuery, countQuery := s.buildSearchPromocodeQuery(filter)
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
			"failed to count promocodes",
			err,
		)
	}

	if totalCount == 0 {
		return []models.Promocode{}, 0, errx.NewNotFound().WithDescription("no promocodes found")
	}

	// Use squirrelHelper to execute the base query
	rows, err := s.squirrelHelper.Query(ctx, s.GetQuerier(), baseQuery)
	if err != nil {
		return nil, 0, errx.NewInternal().WithDescriptionAndCause(
			"failed to query promocodes",
			err,
		)
	}
	defer rows.Close()

	promocodes, err := s.scanPromocodes(rows)
	if err != nil {
		return nil, 0, err
	}

	return promocodes, totalCount, nil
}

func (s *Storage) buildSearchPromocodeQuery(filter dto.ListPromocodeFilter) (squirrel.SelectBuilder, squirrel.SelectBuilder) {
	baseQuery := s.Builder().Select(
		"id",
		"code",
		"discount",
		"expires_at",
		"created_at",
		"updated_at",
	).From("promocodes")

	countQuery := s.Builder().Select("COUNT(*)").From("promocodes")

	if filter.ID != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"id": filter.ID})
		countQuery = countQuery.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.Code != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"code": "%" + filter.Code + "%"})
		countQuery = countQuery.Where(squirrel.ILike{"code": "%" + filter.Code + "%"})
	}
	if filter.DiscountFrom > 0 {
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"discount": filter.DiscountFrom})
		countQuery = countQuery.Where(squirrel.GtOrEq{"discount": filter.DiscountFrom})
	}
	if filter.DiscountTo > 0 {
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"discount": filter.DiscountTo})
		countQuery = countQuery.Where(squirrel.LtOrEq{"discount": filter.DiscountTo})
	}
	if filter.Active {
		now := time.Now()
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"expires_at": now})
		countQuery = countQuery.Where(squirrel.GtOrEq{"expires_at": now})
	}
	if filter.Expired {
		now := time.Now()
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"expires_at": now})
		countQuery = countQuery.Where(squirrel.LtOrEq{"expires_at": now})
	}

	return baseQuery, countQuery
}

func (s *Storage) scanPromocodes(rows pgx.Rows) ([]models.Promocode, error) {
	var promocodes []models.Promocode

	for rows.Next() {
		var promocode models.Promocode

		err := rows.Scan(
			&promocode.ID,
			&promocode.Code,
			&promocode.Discount,
			&promocode.ExpiresAt,
			&promocode.CreatedAt,
			&promocode.UpdatedAt,
		)

		if err != nil {
			return nil, errx.NewInternal().WithDescriptionAndCause(
				"failed to scan promocode",
				err,
			)
		}

		promocodes = append(promocodes, promocode)
	}

	if err := rows.Err(); err != nil {
		return nil, errx.NewInternal().WithDescriptionAndCause(
			"rows error",
			err,
		)
	}

	return promocodes, nil
}

func (s *Storage) DeletePromocode(ctx context.Context, id string) error {
	result, err := s.GetQuerier().Exec(ctx, "DELETE FROM promocodes WHERE id = $1", id)
	if err != nil {
		return errx.NewInternal().WithDescriptionAndCause(
			"promocode deletion failed",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return errx.NewNotFound().WithDescription(fmt.Sprintf("promocode with id '%s' not found", id))
	}

	return nil
}
