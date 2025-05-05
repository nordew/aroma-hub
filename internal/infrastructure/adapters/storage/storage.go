package storage

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nordew/go-errx"
	pgxtransactor "github.com/nordew/pgx-transactor"
)

const (
	uniqueViolationCode = "23505"
)

type Storage struct {
	pgxtransactor.Storage
	squirrelHelper *pgxtransactor.SquirrelHelper
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		Storage:        pgxtransactor.NewBaseStorage(pool),
		squirrelHelper: pgxtransactor.NewSquirrelHelper(),
	}
}

func handleSQLError(err error, entity, id string) error {
	if strings.Contains(err.Error(), "duplicate key") {
		return errx.NewAlreadyExists().WithDescriptionAndCause(
			fmt.Sprintf("%s with id '%s' already exists", entity, id),
			err,
		)
	}

	return errx.NewInternal().WithDescriptionAndCause(
		fmt.Sprintf("failed to create %s", entity),
		err,
	)
}
