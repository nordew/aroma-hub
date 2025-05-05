package minio_s3

import (
	"aroma-hub/internal/config"
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func MustConnect(cfg config.Minio) *minio.Client {
	host := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "http://"), "https://")
	address := net.JoinHostPort(host, fmt.Sprint(cfg.Port))

	log.Printf("→ dialing MinIO at %q (secure=%v)", address, cfg.UseSSL)
	client, err := minio.New(address, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("failed to connect to MinIO at %q: %v", address, err)
	}

	if cfg.BucketName == "" {
		log.Fatalf("no bucket name configured")
	}

	exists, err := client.BucketExists(context.Background(), cfg.BucketName)
	if err != nil {
		log.Fatalf("failed to check bucket %q: %v", cfg.BucketName, err)
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("failed to create bucket %q: %v", cfg.BucketName, err)
		}
		log.Printf("created bucket %q", cfg.BucketName)
	}

	return client
}
