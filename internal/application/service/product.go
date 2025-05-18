package service

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/nordew/go-errx"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	presignExpiry = 280 * time.Minute
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

func (s *Service) ListProducts(
	ctx context.Context,
	filter dto.ListProductFilter,
) (dto.ListProductResponse, error) {
	products, total, err := s.storage.ListProducts(ctx, filter)
	if err != nil {
		return dto.ListProductResponse{}, errors.Wrap(err, "ListProducts: failed to list products")
	}

	resp := dto.ListProductResponse{
		Count:    total,
		Products: make([]models.Product, 0, len(products)),
	}

	for _, p := range products {
		prod := p

		filename, err := s.getLatestImageName(ctx, p.ID)
		if err == nil && filename != "" {
			url, err := s.presignGetObject(ctx, p.ID, filename)
			if err == nil {
				prod.ImageURL = url
			}
		}

		resp.Products = append(resp.Products, prod)
	}

	return resp, nil
}

func (s *Service) getLatestImageName(
	ctx context.Context,
	productID string,
) (string, error) {
	objectCh := s.minioClient.ListObjects(ctx, s.minioBucket, minio.ListObjectsOptions{
		Prefix:    productID + "/",
		Recursive: true,
	})

	var (
		latestName string
		latestTime time.Time
	)

	for obj := range objectCh {
		if obj.Err != nil {
			continue
		}

		if obj.LastModified.After(latestTime) {
			latestTime = obj.LastModified
			latestName = path.Base(obj.Key)
		}
	}

	return latestName, nil
}

func (s *Service) presignGetObject(
	ctx context.Context,
	productID,
	filename string,
) (string, error) {
	objectPath := path.Join(productID, filename)

	url, err := s.minioClient.PresignedGetObject(ctx, s.minioBucket, objectPath, presignExpiry, nil)
	if err != nil {
		return "", errx.NewInternal().WithDescription("failed to generate presigned URL")
	}

	return url.String(), nil
}

func (s *Service) ListBrands(ctx context.Context) (dto.BrandResponse, error) {
	brands, err := s.storage.ListBrands(ctx)
	if err != nil {
		return dto.BrandResponse{}, err
	}

	resp := dto.BrandResponse{
		Brands: brands,
	}

	return resp, nil
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

func (s *Service) SetProductImage(
	ctx context.Context,
	productID string,
	imageData []byte,
) error {
	if len(imageData) == 0 {
		return nil
	}

	_, err := s.uploadProductImage(ctx, productID, imageData)
	return err
}

func (s *Service) uploadProductImage(
	ctx context.Context,
	productID string,
	imageData []byte,
) (string, error) {
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", errx.NewBadRequest().WithDescription("uploaded data is not a valid image")
	}

	ext := format
	if ext == "jpeg" {
		ext = "jpg"
	}

	buf := &bytes.Buffer{}

	switch ext {
	case "png":
		err = png.Encode(buf, img)

	case "jpg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})

	default:
		return "", errx.NewBadRequest().WithDescription(
			fmt.Sprintf("unsupported image format: %s", format),
		)
	}

	if err != nil {
		return "", errx.NewInternal().WithDescription("failed to encode image for upload")
	}

	filename := uuid.NewString() + "." + ext
	objectPath := path.Join(productID, filename)
	size := int64(buf.Len())

	_, err = s.minioClient.PutObject(
		ctx,
		s.minioBucket,
		objectPath,
		buf,
		size,
		minio.PutObjectOptions{ContentType: fmt.Sprintf("image/%s", ext)},
	)
	if err != nil {
		return "", errx.NewInternal().WithDescription("failed to upload image to object storage")
	}

	return filename, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	return s.storage.DeleteProduct(ctx, id)
}
