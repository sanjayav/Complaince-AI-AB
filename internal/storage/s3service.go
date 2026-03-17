package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// S3Service handles all S3 operations for document management
type S3Service struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
	region     string
}

// DocumentInfo represents metadata about a stored document
type DocumentInfo struct {
	Key         string            `json:"key"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	UploadedAt  time.Time         `json:"uploaded_at"`
	ETag        string            `json:"etag"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewS3Service creates a new S3 service
func NewS3Service(bucketName, region string) (*S3Service, error) {
	// Create a new client with the same config as the signer
	cfg, err := loadAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(client)

	return &S3Service{
		client:     client,
		presigner:  presigner,
		bucketName: bucketName,
		region:     region,
	}, nil
}

// UploadDocument uploads a document to S3
func (s *S3Service) UploadDocument(ctx context.Context, file *multipart.FileHeader, docType string) (*DocumentInfo, error) {
	// Generate unique document ID
	docID := uuid.New().String()

	// Determine file extension
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".pdf" // Default to PDF if no extension
	}

	// Create S3 key
	key := fmt.Sprintf("documents/%s/%s%s", docType, docID, ext)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(key),
		Body:          src,
		ContentType:   aws.String(file.Header.Get("Content-Type")),
		ContentLength: aws.Int64(file.Size),
		Metadata: map[string]string{
			"original-filename": file.Filename,
			"uploaded-by":       "system",
			"document-type":     docType,
			"upload-timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload document to S3: %w", err)
	}

	log.Printf("Successfully uploaded document: %s (size: %d bytes)", key, file.Size)

	return &DocumentInfo{
		Key:         key,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
		UploadedAt:  time.Now().UTC(),
		ETag:        "", // Would be set from PutObject response in production
		Metadata: map[string]string{
			"original-filename": file.Filename,
			"document-type":     docType,
		},
	}, nil
}

// UploadDocumentFromReader uploads a document to S3 from an io.Reader
func (s *S3Service) UploadDocumentFromReader(ctx context.Context, filename string, content io.Reader, size int64, contentType, docType string) (*DocumentInfo, error) {
	// Generate unique document ID
	docID := uuid.New().String()

	// Determine file extension
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".pdf" // Default to PDF if no extension
	}

	// Create S3 key
	key := fmt.Sprintf("documents/%s/%s%s", docType, docID, ext)

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(key),
		Body:          content,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
		Metadata: map[string]string{
			"original-filename": filename,
			"uploaded-by":       "system",
			"document-type":     docType,
			"upload-timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload document to S3: %w", err)
	}

	log.Printf("Successfully uploaded document: %s (size: %d bytes)", key, size)

	return &DocumentInfo{
		Key:         key,
		Size:        size,
		ContentType: contentType,
		UploadedAt:  time.Now().UTC(),
		ETag:        "", // Would be set from PutObject response in production
		Metadata: map[string]string{
			"original-filename": filename,
			"document-type":     docType,
		},
	}, nil
}

// DownloadDocument downloads a document from S3
func (s *S3Service) DownloadDocument(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download document: %w", err)
	}

	return result.Body, nil
}

// ListDocuments lists documents in a specific prefix
func (s *S3Service) ListDocuments(ctx context.Context, prefix string, maxKeys int32) ([]DocumentInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucketName),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(maxKeys),
	}

	result, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	var documents []DocumentInfo
	for _, obj := range result.Contents {
		doc := DocumentInfo{
			Key:        *obj.Key,
			Size:       *obj.Size,
			UploadedAt: *obj.LastModified,
			ETag:       strings.Trim(*obj.ETag, `"`),
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// DeleteDocument deletes a document from S3
func (s *S3Service) DeleteDocument(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	log.Printf("Successfully deleted document: %s", key)
	return nil
}

// GetDocumentInfo gets metadata about a document
func (s *S3Service) GetDocumentInfo(ctx context.Context, key string) (*DocumentInfo, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get document info: %w", err)
	}

	return &DocumentInfo{
		Key:         key,
		Size:        *result.ContentLength,
		ContentType: aws.ToString(result.ContentType),
		UploadedAt:  *result.LastModified,
		ETag:        strings.Trim(aws.ToString(result.ETag), `"`),
		Metadata:    result.Metadata,
	}, nil
}

// CreateBucket creates the S3 bucket if it doesn't exist
func (s *S3Service) CreateBucket(ctx context.Context) error {
	_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s.region),
		},
	})

	if err != nil {
		// Check if bucket already exists
		var bucketAlreadyExists *types.BucketAlreadyExists
		var bucketAlreadyOwnedByYou *types.BucketAlreadyOwnedByYou
		if !(errors.As(err, &bucketAlreadyExists) || errors.As(err, &bucketAlreadyOwnedByYou)) {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Bucket %s already exists", s.bucketName)
	} else {
		log.Printf("Successfully created bucket: %s", s.bucketName)
	}

	return nil
}

// BucketExists checks if the bucket exists
func (s *S3Service) BucketExists(ctx context.Context) (bool, error) {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})

	if err != nil {
		var noSuchBucket *types.NoSuchBucket
		if errors.As(err, &noSuchBucket) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	return true, nil
}

// GetPresignedURL generates a presigned URL for document access
func (s *S3Service) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	request, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// UploadPresignedURL generates a presigned URL for document upload
func (s *S3Service) UploadPresignedURL(ctx context.Context, key, contentType string, expires time.Duration) (string, error) {
	request, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate upload presigned URL: %w", err)
	}

	return request.URL, nil
}

// Close cleans up resources
func (s *S3Service) Close() error {
	// S3 client doesn't need explicit cleanup
	return nil
}
