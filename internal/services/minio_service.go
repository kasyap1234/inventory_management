package services

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/cors"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioService interface {
	UploadImage(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) error
	UploadObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error
	GetPresignedURL(bucketName, objectName string, expiry time.Duration) (string, error)
	GetPresignedPutURL(ctx context.Context, bucketName, objectName string, expiry time.Duration, contentType string) (string, error)
	FetchObject(ctx context.Context, bucketName, objectName string) ([]byte, string, error)
	DeleteImage(ctx context.Context, bucketName, objectName string) error
	EnsureBucketExists(ctx context.Context, bucketName string) error
	HealthCheck(ctx context.Context) error
	SetupBucketsWithCORS(ctx context.Context, allowedOrigins []string) error
}

type minioClient struct {
	client *minio.Client
}

func NewMinioService(endpoint, accessKey, secretKey string, useSSL bool) (MinioService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &minioClient{client: client}, nil
}

func (m *minioClient) UploadImage(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) error {
	return m.UploadObject(ctx, bucketName, objectName, reader, objectSize, "image/jpeg")
}

func (m *minioClient) UploadObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if opts.ContentType == "" {
		opts.ContentType = "application/octet-stream"
	}
	_, err := m.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
	return err
}

func (m *minioClient) GetPresignedURL(bucketName, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(context.Background(), bucketName, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (m *minioClient) DeleteImage(ctx context.Context, bucketName, objectName string) error {
	return m.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

func (m *minioClient) GetPresignedPutURL(ctx context.Context, bucketName, objectName string, expiry time.Duration, contentType string) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, bucketName, objectName, expiry)
	if err != nil {
		return "", err
	}
	if contentType != "" {
		// Append content-type so browsers send it during PUT
		q := url.Query()
		q.Set("Content-Type", contentType)
		url.RawQuery = q.Encode()
	}
	return url.String(), nil
}

func (m *minioClient) FetchObject(ctx context.Context, bucketName, objectName string) ([]byte, string, error) {
	obj, err := m.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	defer obj.Close()

	info, statErr := obj.Stat()
	data, readErr := io.ReadAll(obj)
	contentType := ""
	if statErr == nil {
		contentType = info.ContentType
	}
	if readErr != nil {
		return nil, contentType, readErr
	}
	return data, contentType, nil
}

func (m *minioClient) EnsureBucketExists(ctx context.Context, bucketName string) error {
	found, err := m.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !found {
		return m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

func (m *minioClient) HealthCheck(ctx context.Context) error {
	// Try to list buckets to verify connectivity
	_, err := m.client.ListBuckets(ctx)
	return err
}

// SetupBucketsWithCORS creates required buckets and configures CORS for browser direct uploads.
// This should be called at application startup to ensure buckets are ready.
func (m *minioClient) SetupBucketsWithCORS(ctx context.Context, allowedOrigins []string) error {
	buckets := []string{"product-images", "invoices"}

	// Default allowed origins for local development
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
		}
	}

	// CORS configuration for browser direct uploads
	corsConfig := &cors.Config{
		CORSRules: []cors.Rule{
			{
				AllowedOrigin: allowedOrigins,
				AllowedMethod: []string{"GET", "PUT", "POST", "DELETE", "HEAD"},
				AllowedHeader: []string{"*"},
				ExposeHeader:  []string{"ETag", "Content-Length", "Content-Type", "x-amz-meta-*"},
				MaxAgeSeconds: 3600,
			},
		},
	}

	for _, bucket := range buckets {
		// Create bucket if it doesn't exist
		if err := m.EnsureBucketExists(ctx, bucket); err != nil {
			log.Printf("Warning: Failed to create bucket %s: %v", bucket, err)
			continue
		}

		// Set CORS configuration
		if err := m.client.SetBucketCors(ctx, bucket, corsConfig); err != nil {
			log.Printf("Warning: Failed to set CORS for bucket %s: %v", bucket, err)
			continue
		}

		log.Printf("✓ Bucket %s configured with CORS", bucket)
	}

	return nil
}
