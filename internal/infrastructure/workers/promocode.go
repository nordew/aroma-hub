package workers

import (
	"aroma-hub/internal/application/dto"
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	midnight = "0 0 * * *"

	// Only for testing purposes
	every30Sec = "*/30 * * * * *"
)

type Service interface {
	ListPromocodes(ctx context.Context, filter dto.ListPromocodeFilter) (dto.ListPromocodesResponse, error)
	DeleteExpiredPromocodes(ctx context.Context) (int64, error)
}

type PromocodeWorker struct {
	cron    *cron.Cron
	service Service
	logger  *log.Logger
}

func NewPromocodeWorker(service Service, logger *log.Logger) *PromocodeWorker {
	cronOptions := cron.WithParser(
		cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		),
	)

	return &PromocodeWorker{
		cron:    cron.New(cronOptions),
		service: service,
		logger:  logger,
	}
}

func (w *PromocodeWorker) Start() {
	_, err := w.cron.AddFunc(midnight, w.cleanExpiredPromocodes)
	if err != nil {
		w.logger.Printf("Failed to schedule promocode cleanup job: %v", err)
	}

	w.cron.Start()
	w.logger.Println("Promocode worker started successfully")
}

func (w *PromocodeWorker) Stop() {
	w.logger.Println("Stopping promocode worker...")

	ctx := w.cron.Stop()
	<-ctx.Done()

	w.logger.Println("Promocode worker stopped successfully")
}

func (w *PromocodeWorker) cleanExpiredPromocodes() {
	w.logger.Println("Starting expired promocode cleanup job")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := w.service.DeleteExpiredPromocodes(ctx)
	if err != nil {
		w.logger.Printf("Error cleaning up expired promocodes: %v", err)
		return
	}
}
