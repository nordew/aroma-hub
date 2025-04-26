package storage

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/nordew/go-errx"
)

var (
	ErrFailedToScanAdmin   = "failed to scan admin"
	ErrFailedToQueryAdmins = "failed to query admins"
)

func (s *Storage) ListAdmins(ctx context.Context, filter dto.ListAdminFilter) ([]models.Admin, error) {
	query := s.buildSearchAdminQuery(ctx, filter)

	rows, err := s.squirrelHelper.Query(ctx, s.GetQuerier(), query)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errx.NewNotFound().WithDescriptionAndCause(
				ErrFailedToQueryAdmins,
				err,
			)
		}

		return nil, errx.NewInternal().WithDescriptionAndCause(
			ErrFailedToQueryAdmins,
			err,
		)
	}
	defer rows.Close()

	return s.scanAdmins(rows)
}

func (s *Storage) buildSearchAdminQuery(_ context.Context, filter dto.ListAdminFilter) squirrel.SelectBuilder {
	query := s.Builder().Select(
		"id",
		"vendor_id",
		"vendor_type",
		"created_at",
		"updated_at",
	).From("admins")

	if filter.VendorID != "" {
		query = query.Where(squirrel.Eq{"vendor_id": filter.VendorID})
	}

	return query
}

func (s *Storage) scanAdmins(rows pgx.Rows) ([]models.Admin, error) {
	var admins []models.Admin

	for rows.Next() {
		var admin models.Admin
		err := rows.Scan(
			&admin.ID,
			&admin.VendorID,
			&admin.VendorType,
			&admin.CreatedAt,
			&admin.UpdatedAt,
		)

		if err != nil {
			return nil, errx.NewInternal().WithDescriptionAndCause(
				ErrFailedToScanAdmin,
				err,
			)
		}

		admins = append(admins, admin)
	}

	if err := rows.Err(); err != nil {
		return nil, errx.NewInternal().WithDescriptionAndCause(
			ErrRowsError,
			err,
		)
	}

	return admins, nil
}
