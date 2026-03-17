package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// S3Event represents an S3 event notification
type S3Event struct {
	Records []S3Record `json:"Records"`
}

// S3Record represents a single S3 event record
type S3Record struct {
	EventName string `json:"eventName"`
	S3        S3Data `json:"s3"`
}

// S3Data contains S3 object information
type S3Data struct {
	Bucket S3Bucket `json:"bucket"`
	Object S3Object `json:"object"`
}

// S3Bucket contains bucket information
type S3Bucket struct {
	Name string `json:"name"`
}

// S3Object contains object information
type S3Object struct {
	Key       string            `json:"key"`
	Size      int64             `json:"size"`
	ETag      string            `json:"eTag"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"createdAt,omitempty"`
}

// S3WebhookHandler handles S3 webhook events for automatic indexing
type S3WebhookHandler struct {
	workerPool *WorkerPool
	jobQueue   chan Job
}

// NewS3WebhookHandler creates a new S3 webhook handler
func NewS3WebhookHandler(workerPool *WorkerPool) *S3WebhookHandler {
	return &S3WebhookHandler{
		workerPool: workerPool,
		jobQueue:   make(chan Job, 100),
	}
}

// HandleWebhook processes S3 webhook events
func (h *S3WebhookHandler) HandleWebhook(ctx context.Context, eventData []byte) error {
	var event S3Event
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to unmarshal S3 event: %w", err)
	}

	log.Printf("Processing S3 webhook with %d records", len(event.Records))

	for _, record := range event.Records {
		if err := h.processRecord(ctx, record); err != nil {
			log.Printf("Error processing S3 record: %v", err)
			continue
		}
	}

	return nil
}

// processRecord processes a single S3 record
func (h *S3WebhookHandler) processRecord(ctx context.Context, record S3Record) error {
	// Only process object creation events
	if record.EventName != "ObjectCreated:Put" && record.EventName != "ObjectCreated:Post" {
		log.Printf("Skipping non-creation event: %s", record.EventName)
		return nil
	}

	// Check if it's a PDF document
	if !isPDFDocument(record.S3.Object.Key) {
		log.Printf("Skipping non-PDF document: %s", record.S3.Object.Key)
		return nil
	}

	// Extract document type from metadata or key
	docType := extractDocumentType(record.S3.Object.Key, record.S3.Object.Metadata)

	// Create processing job
	job := Job{
		ID:       uuid.New().String(),
		S3Key:    record.S3.Object.Key,
		DocType:  docType,
		Priority: getJobPriority(docType),
		Created:  time.Now(),
	}

	log.Printf("Creating processing job for S3 object: %s (type: %s)", record.S3.Object.Key, docType)

	// Submit to worker pool
	if err := h.workerPool.SubmitJob(job); err != nil {
		return fmt.Errorf("failed to submit job: %w", err)
	}

	return nil
}

// isPDFDocument checks if the S3 key represents a PDF document
func isPDFDocument(key string) bool {
	// Check file extension
	if len(key) < 4 {
		return false
	}

	ext := key[len(key)-4:]
	return ext == ".pdf" || ext == ".PDF"
}

// extractDocumentType extracts the document type from S3 key or metadata
func extractDocumentType(key string, metadata map[string]string) string {
	// Check metadata first
	if docType, ok := metadata["document-type"]; ok {
		return docType
	}

	// Extract from S3 key path
	if len(key) > 0 {
		// Key format: documents/{type}/{filename}
		parts := strings.Split(key, "/")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// Default to automotive test
	return "automotive_test"
}

// getJobPriority determines job priority based on document type
func getJobPriority(docType string) int {
	switch docType {
	case "safety_critical":
		return 1 // Highest priority
	case "regulatory":
		return 2
	case "automotive_test":
		return 3
	case "research":
		return 4
	default:
		return 5 // Lowest priority
	}
}

// Start starts the webhook handler
func (h *S3WebhookHandler) Start(ctx context.Context) {
	log.Println("Starting S3 webhook handler")

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("S3 webhook handler stopped")
				return
			}
		}
	}()
}

// Stop stops the webhook handler
func (h *S3WebhookHandler) Stop() {
	log.Println("Stopping S3 webhook handler")
	close(h.jobQueue)
}
