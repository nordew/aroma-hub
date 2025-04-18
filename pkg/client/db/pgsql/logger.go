package pgsql

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
	"time"
)

type QueryLogger struct {
	logger *log.Logger
}

func NewQueryLogger() *QueryLogger {
	return &QueryLogger{
		logger: log.New(os.Stdout, "[SQL] ", log.LstdFlags),
	}
}

func (l *QueryLogger) TraceQueryStart(
	ctx context.Context,
	conn *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	return context.WithValue(ctx, "query_start_time", time.Now())
}

func (l *QueryLogger) TraceQueryEnd(
	ctx context.Context,
	conn *pgx.Conn,
	data pgx.TraceQueryEndData,
) {
	startTime, ok := ctx.Value("query_start_time").(time.Time)
	if !ok {
		startTime = time.Now()
	}

	duration := time.Since(startTime)

	query := data

	l.logger.Printf("Query: %s | Params: %v | Duration: %s",
		query, data.CommandTag, duration)

	if data.Err != nil {
		l.logger.Printf("Query Error: %v", data.Err)
	}
}
