package services

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ownafarm/ownafarm-backend/internal/config"
)

// StorageService interface untuk abstraksi storage operations
type StorageService interface {
	// Upload file ke storage
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error

	// Download file dari storage
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// GetPresignedUploadURL generates presigned URL untuk client-side upload
	GetPresignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error)

	// GetPresignedDownloadURL generates presigned URL untuk download
	GetPresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// R2StorageService implementasi StorageService untuk Cloudflare R2
type R2StorageService struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
}

// NewR2StorageService creates a new R2StorageService instance
func NewR2StorageService(cfg *config.R2Config) (*R2StorageService, error) {
	// Create custom resolver for R2 endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: cfg.Endpoint,
		}, nil
	})

	// Create AWS config with static credentials
	awsCfg := aws.Config{
		Region:                      cfg.Region,
		Credentials:                 credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		EndpointResolverWithOptions: customResolver,
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	// Create presign client
	presigner := s3.NewPresignClient(client)

	return &R2StorageService{
		client:    client,
		presigner: presigner,
		bucket:    cfg.Bucket,
	}, nil
}

// Upload uploads a file to R2
func (s *R2StorageService) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

// Download downloads a file from R2
func (s *R2StorageService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

// GetPresignedUploadURL generates a presigned URL for uploading
func (s *R2StorageService) GetPresignedUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	presignedReq, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return presignedReq.URL, nil
}

// GetPresignedDownloadURL generates a presigned URL for downloading
func (s *R2StorageService) GetPresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignedReq, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return presignedReq.URL, nil
}
