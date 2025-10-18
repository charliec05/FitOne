package storage

import (
	"context"
	"fmt"
	"time"

	"fitonex/backend/internal/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Service handles S3-compatible storage operations
type S3Service struct {
	client *s3.S3
	bucket string
	region string
}

// NewS3Service creates a new S3 service
func NewS3Service(cfg *config.Config) (*S3Service, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.S3Region),
		Endpoint:         aws.String(cfg.S3Endpoint),
		S3ForcePathStyle: aws.Bool(true), // Required for MinIO
		Credentials:      credentials.NewStaticCredentials("minioadmin", "minioadmin", ""), // MinIO defaults
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := s3.New(sess)

	return &S3Service{
		client: client,
		bucket: cfg.S3Bucket,
		region: cfg.S3Region,
	}, nil
}

// PresignPut generates a presigned URL for uploading a file
func (s *S3Service) PresignPut(ctx context.Context, key, contentType string, sizeBytes int64, ttl time.Duration) (string, error) {
	req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})

	// Set content length if provided
	if sizeBytes > 0 {
		req.HTTPRequest.Header.Set("Content-Length", fmt.Sprintf("%d", sizeBytes))
	}

	url, err := req.Presign(ttl)
	if err != nil {
		return "", fmt.Errorf("failed to presign PUT request: %w", err)
	}

	return url, nil
}

// PresignGet generates a presigned URL for downloading a file
func (s *S3Service) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(ttl)
	if err != nil {
		return "", fmt.Errorf("failed to presign GET request: %w", err)
	}

	return url, nil
}

// DeleteObject deletes an object from S3
func (s *S3Service) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// ObjectExists checks if an object exists in S3
func (s *S3Service) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		if s.client.IsAPIError(err, "NoSuchKey") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}
