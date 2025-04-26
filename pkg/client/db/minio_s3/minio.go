package minio_s3

import (
	"aroma-hub/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	connectTimeout = 30 * time.Second
	retryAttempts  = 5
	retryDelay     = 2 * time.Second
)

type Client struct {
	*minio.Client
	Bucket string
}

func MustConnect(cfg config.Minio) *Client {
	client, err := Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v (endpoint: %s, bucket: %s)",
			err, cfg.Endpoint, cfg.BucketName)
	}

	return client
}

func Connect(cfg config.Minio) (*Client, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("MinIO access key or secret key not provided")
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("MinIO endpoint not provided")
	}
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("MinIO bucket name not provided")
	}

	var client *minio.Client
	var err error

	for attempt := 1; attempt <= retryAttempts; attempt++ {
		log.Printf("Attempting to connect to MinIO (endpoint: %s, useSSL: %t, attempt: %d/%d)",
			cfg.Endpoint, cfg.UseSSL, attempt, retryAttempts)

		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		defer cancel()

		client, err = minio.New(cfg.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: cfg.UseSSL,
		})

		if err == nil {
			_, err = client.BucketExists(ctx, cfg.BucketName)
			if err == nil {
				break
			}
		}

		log.Printf("Failed to connect to MinIO: %v (attempt: %d/%d)",
			err, attempt, retryAttempts)

		if attempt < retryAttempts {
			backoff := time.Duration(attempt) * retryDelay
			time.Sleep(backoff)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO after %d attempts: %w", retryAttempts, err)
	}

	log.Printf("Successfully connected to MinIO (endpoint: %s)", cfg.Endpoint)

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		log.Printf("Creating bucket: %s", cfg.BucketName)

		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		log.Printf("Bucket created successfully: %s", cfg.BucketName)
	} else {
		log.Printf("Bucket already exists: %s", cfg.BucketName)
	}

	return &Client{
		Client: client,
		Bucket: cfg.BucketName,
	}, nil
}
