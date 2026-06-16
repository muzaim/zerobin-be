package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	UploadAvatar(ctx context.Context, filename string, reader io.Reader, size int64, contentType string) (string, error)
	UploadFile(ctx context.Context, folder string, filename string, reader io.Reader, size int64, contentType string) (string, error)
}

type minioStorage struct {
	client     *minio.Client
	bucketName string
	publicURL  string
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucketName, publicURL string, useSSL bool) (Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)

		err = client.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			return nil, fmt.Errorf("failed to set public read policy on bucket: %w", err)
		}
	}

	if publicURL == "" {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		publicURL = fmt.Sprintf("%s://%s/%s", scheme, endpoint, bucketName)
	}

	return &minioStorage{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}, nil
}

func (s *minioStorage) UploadAvatar(ctx context.Context, filename string, reader io.Reader, size int64, contentType string) (string, error) {
	objectName := "avatars/" + filename
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.publicURL, objectName), nil
}

func (s *minioStorage) UploadFile(ctx context.Context, folder string, filename string, reader io.Reader, size int64, contentType string) (string, error) {
	objectName := folder + "/" + filename
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.publicURL, objectName), nil
}
