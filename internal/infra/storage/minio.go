package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"idsai-core-up/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStorage interface {
	PutObject(ctx context.Context, key, contentType string, body []byte) error
	DeleteObject(ctx context.Context, key string) error
	PublicURL(key string) string
	Available() bool
}

type nopStorage struct{}

func (n nopStorage) PutObject(ctx context.Context, key, contentType string, body []byte) error {
	return fmt.Errorf("object storage unavailable")
}

func (n nopStorage) DeleteObject(ctx context.Context, key string) error {
	return nil
}

func (n nopStorage) PublicURL(key string) string {
	return ""
}

func (n nopStorage) Available() bool {
	return false
}

type minioStorage struct {
	client           *minio.Client
	bucket           string
	publicBaseURL    string
	endpoint         string
	useSSL           bool
	ensureBucketOnce sync.Once
	ensureBucketErr  error
}

func publicReadPolicy(bucket string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket)
}

func NewFromConfig(cfg config.Config) ObjectStorage {
	if strings.TrimSpace(cfg.LocalStorageDir) != "" {
		return newLocalStorage(cfg.LocalStorageDir, cfg.PublicBaseURL)
	}

	endpoint := strings.TrimSpace(cfg.StorageEndpoint)
	accessKey := strings.TrimSpace(cfg.StorageAccessKey)
	secretKey := strings.TrimSpace(cfg.StorageSecretKey)
	bucket := strings.TrimSpace(cfg.StorageBucket)
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return newLocalStorage(cfg.LocalStorageDir, cfg.PublicBaseURL)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: cfg.StorageUseSSL,
	})
	if err != nil {
		return newLocalStorage(cfg.LocalStorageDir, cfg.PublicBaseURL)
	}

	return &minioStorage{
		client:        client,
		bucket:        bucket,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.StoragePublicBaseURL), "/"),
		endpoint:      endpoint,
		useSSL:        cfg.StorageUseSSL,
	}
}

func (s *minioStorage) ensureBucket(ctx context.Context) error {
	s.ensureBucketOnce.Do(func() {
		exists, err := s.client.BucketExists(ctx, s.bucket)
		if err != nil {
			s.ensureBucketErr = err
			return
		}
		if exists {
			s.ensureBucketErr = s.client.SetBucketPolicy(ctx, s.bucket, publicReadPolicy(s.bucket))
			return
		}
		s.ensureBucketErr = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
		if s.ensureBucketErr != nil {
			return
		}
		s.ensureBucketErr = s.client.SetBucketPolicy(ctx, s.bucket, publicReadPolicy(s.bucket))
	})
	return s.ensureBucketErr
}

func (s *minioStorage) PutObject(ctx context.Context, key, contentType string, body []byte) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("empty object key")
	}
	if err := s.ensureBucket(ctx); err != nil {
		return err
	}

	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *minioStorage) DeleteObject(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	if err := s.ensureBucket(ctx); err != nil {
		return err
	}
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *minioStorage) PublicURL(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + s.bucket + "/" + key
	}

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, key)
}

func (s *minioStorage) Available() bool {
	return true
}
