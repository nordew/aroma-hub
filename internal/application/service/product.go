package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/shopspring/decimal"
)

func (s *Service) CreateProduct(ctx context.Context, input dto.CreateProductRequest) error {
	categories, _, err := s.storage.ListCategories(ctx, dto.ListCategoryFilter{
		Name: input.CategoryName,
	})
	if err != nil {
		return err
	}
	category := categories[0]

	product, err := models.NewProduct(
		uuid.NewString(),
		category.ID,
		input.Brand,
		input.Name,
		input.Description,
		input.Composition,
		input.Characteristics,
		decimal.NewFromFloat(input.Price),
		input.StockAmount,
	)
	if err != nil {
		return err
	}

	return s.storage.CreateProduct(ctx, product)
}

func (s *Service) ListProducts(ctx context.Context, filter dto.ListProductFilter) (dto.ListProductResponse, error) {
	products, total, err := s.storage.ListProducts(ctx, filter)
	if err != nil {
		return dto.ListProductResponse{}, fmt.Errorf("failed to list products: %w", err)
	}

	for i := range products {
		if products[i].ImageURL != "" {
			parts := strings.Split(products[i].ImageURL, "/")
			if len(parts) >= 3 {
				objectName := parts[len(parts)-1]

				url, err := s.getImageURL(ctx, objectName)
				if err != nil {
					products[i].ImageURL = ""
					continue
				}

				products[i].ImageURL = url
			}
		}
	}

	return dto.ListProductResponse{
		Products: products,
		Count:    total,
	}, nil
}

func (s *Service) getImageURL(ctx context.Context, objectName string) (string, error) {
	url, err := s.minioCl.PresignedGetObject(
		ctx,
		s.minioCl.Bucket,
		objectName,
		time.Hour*24,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	urlStr := url.String()

	return urlStr, nil
}

func (s *Service) UpdateProduct(ctx context.Context, input dto.UpdateProductRequest) error {
	var newCategoryName string
	if input.CategoryName != "" {
		categories, _, err := s.storage.ListCategories(ctx, dto.ListCategoryFilter{
			Name: input.CategoryName,
		})
		if err != nil {
			return err
		}
		category := categories[0]

		newCategoryName = category.Name
	}

	input.CategoryName = newCategoryName

	return s.storage.UpdateProduct(ctx, input)
}

func (s *Service) SetProductImage(ctx context.Context, productID string, imageBytes []byte) error {
	_, _, err := s.storage.ListProducts(ctx, dto.ListProductFilter{
		IDs:           []string{productID},
		ShowInvisible: true,
	})
	if err != nil {
		return err
	}

	imageFileName := fmt.Sprintf("%s-%s", productID, uuid.New().String())

	imageReader := bytes.NewReader(imageBytes)

	contentType := http.DetectContentType(imageBytes)

	_, err = s.minioCl.PutObject(
		ctx,
		s.minioCl.Bucket,
		imageFileName,
		imageReader,
		int64(len(imageBytes)),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return err
	}

	imageURL := fmt.Sprintf("/%s/%s", s.minioCl.Bucket, imageFileName)

	updateProductReq := dto.UpdateProductRequest{
		ID:       productID,
		ImageURL: imageURL,
	}
	if err := s.storage.UpdateProduct(ctx, updateProductReq); err != nil {
		_ = s.minioCl.RemoveObject(ctx, s.minioCl.Bucket, imageFileName, minio.RemoveObjectOptions{})
		return err
	}

	return nil
}

func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	return s.storage.DeleteProduct(ctx, id)
}
