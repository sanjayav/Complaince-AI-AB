package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// TenantS3Service provides multi-tenant S3 operations
type TenantS3Service struct {
	s3Client *s3.Client
	bucket   string
	region   string
}

// TenantDocument represents a document in the context of a tenant
type TenantDocument struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	Filename     string            `json:"filename"`
	OriginalName string            `json:"original_name"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	Status       string            `json:"status"` // uploaded, processing, processed, failed
	UploadedAt   time.Time         `json:"uploaded_at"`
	ProcessedAt  *time.Time        `json:"processed_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Checksum     string            `json:"checksum,omitempty"`
}

// BatchUploadResult represents the result of a batch upload operation
type BatchUploadResult struct {
	SuccessCount int                `json:"success_count"`
	FailureCount int                `json:"failure_count"`
	Results      []BatchUploadItem  `json:"results"`
	Errors       []BatchUploadError `json:"errors,omitempty"`
	Duration     time.Duration      `json:"duration"`
}

// BatchUploadItem represents a single item in a batch upload
type BatchUploadItem struct {
	Filename   string        `json:"filename"`
	Status     string        `json:"status"` // success, failed
	DocumentID string        `json:"document_id,omitempty"`
	Error      string        `json:"error,omitempty"`
	Size       int64         `json:"size"`
	UploadTime time.Duration `json:"upload_time"`
}

// BatchUploadError represents an error in batch upload
type BatchUploadError struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

// NewTenantS3Service creates a new tenant-aware S3 service
func NewTenantS3Service(s3Client *s3.Client, bucket, region string) *TenantS3Service {
	return &TenantS3Service{
		s3Client: s3Client,
		bucket:   bucket,
		region:   region,
	}
}

// UploadDocument uploads a document for a specific tenant
func (ts *TenantS3Service) UploadDocument(ctx context.Context, tenantID string, file io.Reader, filename string, contentType string, metadata map[string]string) (*TenantDocument, error) {
	// Generate unique document ID
	documentID := generateDocumentID(tenantID)

	// Create tenant-specific key
	key := ts.getTenantKey(tenantID, documentID, filename)

	// Check if S3 client is available
	if ts.s3Client == nil {
		// In development mode, simulate successful upload without S3
		log.Printf("S3 client not available, simulating upload for development")
		log.Printf("Would upload to S3 key: %s in bucket: %s", key, ts.bucket)
	} else {
		// Upload to S3
		_, err := ts.s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(ts.bucket),
			Key:          aws.String(key),
			Body:         file,
			ContentType:  aws.String(contentType),
			Metadata:     metadata,
			Tagging:      aws.String(fmt.Sprintf("tenant=%s&type=document", tenantID)),
			StorageClass: types.StorageClassStandard,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to upload document to S3: %w", err)
		}
	}

	// Get file size (approximate for now)
	size := int64(0)
	if seeker, ok := file.(io.Seeker); ok {
		if pos, err := seeker.Seek(0, io.SeekCurrent); err == nil {
			size = pos
		}
	}

	document := &TenantDocument{
		ID:           documentID,
		TenantID:     tenantID,
		Filename:     filename,
		OriginalName: filename,
		Size:         size,
		ContentType:  contentType,
		Status:       "uploaded",
		UploadedAt:   time.Now(),
		Metadata:     metadata,
		Tags: map[string]string{
			"tenant": tenantID,
			"type":   "document",
		},
	}

	log.Printf("Uploaded document %s for tenant %s", documentID, tenantID)
	return document, nil
}

// BatchUploadDocuments uploads multiple documents concurrently
func (ts *TenantS3Service) BatchUploadDocuments(ctx context.Context, tenantID string, files []BatchFile, maxConcurrent int) *BatchUploadResult {
	startTime := time.Now()
	result := &BatchUploadResult{
		Results: make([]BatchUploadItem, 0, len(files)),
		Errors:  make([]BatchUploadError, 0),
	}

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Process each file
	for _, file := range files {
		wg.Add(1)
		go func(f BatchFile) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			uploadStart := time.Now()

			// Upload document
			document, err := ts.UploadDocument(ctx, tenantID, f.Reader, f.Filename, f.ContentType, f.Metadata)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.FailureCount++
				result.Errors = append(result.Errors, BatchUploadError{
					Filename: f.Filename,
					Error:    err.Error(),
				})

				result.Results = append(result.Results, BatchUploadItem{
					Filename:   f.Filename,
					Status:     "failed",
					Error:      err.Error(),
					Size:       f.Size,
					UploadTime: time.Since(uploadStart),
				})
			} else {
				result.SuccessCount++
				result.Results = append(result.Results, BatchUploadItem{
					Filename:   f.Filename,
					Status:     "success",
					DocumentID: document.ID,
					Size:       f.Size,
					UploadTime: time.Since(uploadStart),
				})
			}
		}(file)
	}

	// Wait for all uploads to complete
	wg.Wait()

	result.Duration = time.Since(startTime)
	log.Printf("Batch upload completed: %d success, %d failures in %v", result.SuccessCount, result.FailureCount, result.Duration)

	return result
}

// ListTenantDocuments lists all documents for a specific tenant
func (ts *TenantS3Service) ListTenantDocuments(ctx context.Context, tenantID string, prefix string, maxKeys int32) ([]TenantDocument, error) {
	var documents []TenantDocument

	// List objects with tenant prefix
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(ts.bucket),
		Prefix:  aws.String(fmt.Sprintf("tenants/%s/documents/", tenantID)),
		MaxKeys: aws.Int32(maxKeys),
	}

	if prefix != "" {
		input.Prefix = aws.String(fmt.Sprintf("tenants/%s/documents/%s", tenantID, prefix))
	}

	paginator := s3.NewListObjectsV2Paginator(ts.s3Client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, object := range page.Contents {
			// Parse document metadata from object
			document := ts.parseDocumentFromObject(object, tenantID)
			if document != nil {
				documents = append(documents, *document)
			}
		}
	}

	return documents, nil
}

// GetDocument retrieves a document by ID
func (ts *TenantS3Service) GetDocument(ctx context.Context, tenantID, documentID string) (*TenantDocument, error) {
	// Try to find the document by listing with prefix
	prefix := fmt.Sprintf("tenants/%s/documents/%s", tenantID, documentID)

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(ts.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1),
	}

	result, err := ts.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	if len(result.Contents) == 0 {
		return nil, fmt.Errorf("document %s not found", documentID)
	}

	object := result.Contents[0]
	document := ts.parseDocumentFromObject(object, tenantID)
	if document == nil {
		return nil, fmt.Errorf("failed to parse document metadata")
	}

	return document, nil
}

// DeleteDocument deletes a document
func (ts *TenantS3Service) DeleteDocument(ctx context.Context, tenantID, documentID string) error {
	// Find the document first
	document, err := ts.GetDocument(ctx, tenantID, documentID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Delete from S3
	key := ts.getTenantKey(tenantID, documentID, document.Filename)
	_, err = ts.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(ts.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete document from S3: %w", err)
	}

	log.Printf("Deleted document %s for tenant %s", documentID, tenantID)
	return nil
}

// GetDocumentPresignedURL generates a presigned URL for document access
func (ts *TenantS3Service) GetDocumentPresignedURL(ctx context.Context, tenantID, documentID string, expiration time.Duration) (string, error) {
	document, err := ts.GetDocument(ctx, tenantID, documentID)
	if err != nil {
		return "", fmt.Errorf("document not found: %w", err)
	}

	key := ts.getTenantKey(tenantID, documentID, document.Filename)

	presignClient := s3.NewPresignClient(ts.s3Client)
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(ts.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// UpdateDocumentStatus updates the processing status of a document
func (ts *TenantS3Service) UpdateDocumentStatus(ctx context.Context, tenantID, documentID, status string, metadata map[string]string) error {
	document, err := ts.GetDocument(ctx, tenantID, documentID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	// Update status and metadata
	document.Status = status
	if metadata != nil {
		for k, v := range metadata {
			document.Metadata[k] = v
		}
	}

	if status == "processed" {
		now := time.Now()
		document.ProcessedAt = &now
	}

	// Update S3 object metadata
	key := ts.getTenantKey(tenantID, documentID, document.Filename)

	// Convert metadata to S3 metadata format
	s3Metadata := make(map[string]string)
	for k, v := range document.Metadata {
		s3Metadata[fmt.Sprintf("x-amz-meta-%s", k)] = v
	}
	s3Metadata["x-amz-meta-status"] = status
	s3Metadata["x-amz-meta-processed-at"] = document.ProcessedAt.Format(time.RFC3339)

	// Copy object with new metadata
	_, err = ts.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:            aws.String(ts.bucket),
		CopySource:        aws.String(fmt.Sprintf("%s/%s", ts.bucket, key)),
		Key:               aws.String(key),
		Metadata:          s3Metadata,
		MetadataDirective: types.MetadataDirectiveReplace,
	})

	if err != nil {
		return fmt.Errorf("failed to update document metadata: %w", err)
	}

	log.Printf("Updated document %s status to %s for tenant %s", documentID, status, tenantID)
	return nil
}

// GetTenantStorageStats returns storage statistics for a tenant
func (ts *TenantS3Service) GetTenantStorageStats(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	documents, err := ts.ListTenantDocuments(ctx, tenantID, "", 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	var totalSize int64
	var documentCount int
	var processingCount int
	var processedCount int
	var failedCount int

	for _, doc := range documents {
		totalSize += doc.Size
		documentCount++

		switch doc.Status {
		case "processing":
			processingCount++
		case "processed":
			processedCount++
		case "failed":
			failedCount++
		}
	}

	stats := map[string]interface{}{
		"tenant_id":        tenantID,
		"total_documents":  documentCount,
		"total_size_bytes": totalSize,
		"total_size_gb":    float64(totalSize) / (1024 * 1024 * 1024),
		"processing":       processingCount,
		"processed":        processedCount,
		"failed":           failedCount,
		"uploaded":         documentCount - processingCount - processedCount - failedCount,
	}

	return stats, nil
}

// Helper functions
func (ts *TenantS3Service) getTenantKey(tenantID, documentID, filename string) string {
	ext := filepath.Ext(filename)
	// Use the S3 prefix from environment if available, otherwise use default
	prefix := os.Getenv("S3_PREFIX")
	if prefix == "" {
		prefix = "ocr_stage/"
	}
	return fmt.Sprintf("%stenants/%s/documents/%s%s", prefix, tenantID, documentID, ext)
}

func (ts *TenantS3Service) parseDocumentFromObject(object types.Object, tenantID string) *TenantDocument {
	// Extract document ID from key
	key := aws.ToString(object.Key)
	parts := strings.Split(key, "/")
	if len(parts) < 4 {
		return nil
	}

	documentID := parts[2]
	filename := parts[len(parts)-1]

	// Parse metadata from object - AWS SDK v2 doesn't expose metadata in ListObjects
	// We'll need to get this separately or use tags
	metadata := make(map[string]string)
	status := "uploaded"
	var processedAt *time.Time

	// For now, we'll use basic parsing - in production you'd want to get full object metadata
	if strings.Contains(key, "processed") {
		status = "processed"
		now := time.Now()
		processedAt = &now
	}

	var size int64
	if object.Size != nil {
		size = *object.Size
	}

	var lastModified time.Time
	if object.LastModified != nil {
		lastModified = *object.LastModified
	} else {
		lastModified = time.Now()
	}

	return &TenantDocument{
		ID:           documentID,
		TenantID:     tenantID,
		Filename:     filename,
		OriginalName: filename,
		Size:         size,
		ContentType:  "application/octet-stream", // Default content type
		Status:       status,
		UploadedAt:   lastModified,
		ProcessedAt:  processedAt,
		Metadata:     metadata,
		Tags: map[string]string{
			"tenant": tenantID,
			"type":   "document",
		},
	}
}

func generateDocumentID(tenantID string) string {
	return fmt.Sprintf("%s_%s", strings.Replace(tenantID, "tenant_", "", 1), uuid.New().String()[:8])
}

// BatchFile represents a file to be uploaded in batch
type BatchFile struct {
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
	Metadata    map[string]string
}
