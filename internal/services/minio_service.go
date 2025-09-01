package services

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioService interface {
	UploadImage(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) error
	GetPresignedURL(bucketName, objectName string, expiry time.Duration) (string, error)
	DeleteImage(ctx context.Context, bucketName, objectName string) error
	EnsureBucketExists(ctx context.Context, bucketName string) error
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
	_, err := m.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: "image/jpeg", // Assume JPEG, but can detect
	})
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