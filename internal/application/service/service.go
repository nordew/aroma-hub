package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"aroma-hub/pkg/auth"
	"context"

	"github.com/minio/minio-go/v7"
	stash "github.com/nordew/go-stash"

	pgxtransactor "github.com/nordew/pgx-transactor"
)

type Storage interface {
	pgxtransactor.Storage

	CreateProduct(ctx context.Context, product models.Product) error
	ListProducts(ctx context.Context, filter dto.ListProductFilter) ([]models.Product, int64, error)
	ListBrands(ctx context.Context) ([]string, error)
	UpdateProduct(ctx context.Context, input dto.UpdateProductRequest) error
	DeleteProduct(ctx context.Context, id string) error

	CreateCategory(ctx context.Context, category models.Category) error
	ListCategories(ctx context.Context, filter dto.ListCategoryFilter) ([]models.Category, int64, error)
	DeleteCategory(ctx context.Context, id string) error

	CreateOrder(ctx context.Context, order models.Order) (models.Order, error)
	ListOrders(ctx context.Context, filter dto.ListOrderFilter) ([]models.Order, int64, error)
	UpdateOrder(ctx context.Context, input dto.UpdateOrderRequest) error
	DeleteOrder(ctx context.Context, id string) error

	CreateOrderProduct(ctx context.Context, orderProduct models.OrderProduct) error
	ListOrderProducts(ctx context.Context, filter dto.ListOrderProductFilter) ([]models.OrderProduct, int64, error)

	CreatePromocode(ctx context.Context, promocode models.Promocode) error
	ListPromocodes(ctx context.Context, filter dto.ListPromocodeFilter) ([]models.Promocode, int64, error)
	DeletePromocode(ctx context.Context, id string) error

	ListAdmins(ctx context.Context, filter dto.ListAdminFilter) ([]models.Admin, error)
}

type MessagingProvider interface {
	BroadcastMessage(ctx context.Context, text string) error
}

type Service struct {
	storage           Storage
	transactor        *pgxtransactor.Transactor
	cache             stash.Cache
	tokenService      *auth.TokenService
	messagingProvider MessagingProvider
	minioClient       *minio.Client
	minioBucket       string
}

func NewService(
	storage Storage,
	transactor *pgxtransactor.Transactor,
	cache stash.Cache,
	tokenService *auth.TokenService,
	messagingProvider MessagingProvider,
	minioClient *minio.Client,
	minioBucket string,
) *Service {
	return &Service{
		storage:           storage,
		transactor:        transactor,
		cache:             cache,
		tokenService:      tokenService,
		messagingProvider: messagingProvider,
		minioClient:       minioClient,
		minioBucket:       minioBucket,
	}
}
