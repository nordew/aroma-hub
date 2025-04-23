package pgsql

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type pgxTracer struct{}

func (t *pgxTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	log.Printf("SQL Query: %s, Args: %v", data.SQL, data.Args)
	return ctx
}

func (t *pgxTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		log.Printf("SQL Error: %v", data.Err)
	} else {
		log.Printf("SQL Result: Rows affected: %d", data.CommandTag.RowsAffected())
	}
}
