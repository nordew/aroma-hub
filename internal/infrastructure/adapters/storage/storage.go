package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"
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
