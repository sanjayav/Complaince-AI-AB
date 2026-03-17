package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// QdrantClient handles communication with Qdrant vector database
type QdrantClient struct {
	client  *resty.Client
	baseURL string
}

// NewQdrantClient creates a new Qdrant client
func NewQdrantClient(baseURL string) *QdrantClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second)

	return &QdrantClient{
		client:  client,
		baseURL: baseURL,
	}
}

// Collection represents a Qdrant collection
type Collection struct {
	Name       string            `json:"name"`
	Vectors    CollectionVectors `json:"vectors"`
	OnDisk     bool              `json:"on_disk"`
	HnswConfig *HnswConfig       `json:"hnsw_config,omitempty"`
}

type CollectionVectors struct {
	Size     int    `json:"size"`
	Distance string `json:"distance"`
}

type HnswConfig struct {
	M                  int `json:"m"`
	EfConstruct        int `json:"ef_construct"`
	FullScanThreshold  int `json:"full_scan_threshold"`
	MaxIndexingThreads int `json:"max_indexing_threads"`
}

// Point represents a vector point in Qdrant
type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// SearchRequest represents a search query
type SearchRequest struct {
	Vector      []float32              `json:"vector"`
	Limit       int                    `json:"limit"`
	WithPayload bool                   `json:"with_payload"`
	WithVector  bool                   `json:"with_vector"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
	Vector  []float32              `json:"vector,omitempty"`
}

// CreateCollection creates a new collection in Qdrant
func (q *QdrantClient) CreateCollection(ctx context.Context, name string, vectorSize int) error {
	collection := Collection{
		Name: name,
		Vectors: CollectionVectors{
			Size:     vectorSize,
			Distance: "Cosine",
		},
		OnDisk: true,
		HnswConfig: &HnswConfig{
			M:                  16,
			EfConstruct:        100,
			FullScanThreshold:  10000,
			MaxIndexingThreads: 4,
		},
	}

	resp, err := q.client.R().
		SetContext(ctx).
		SetBody(collection).
		Put(fmt.Sprintf("/collections/%s", name))

	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to create collection, status: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	return nil
}

// UpsertPoints inserts or updates points in a collection
func (q *QdrantClient) UpsertPoints(ctx context.Context, collectionName string, points []Point) error {
	payload := map[string]interface{}{
		"points": points,
	}

	resp, err := q.client.R().
		SetContext(ctx).
		SetBody(payload).
		Put(fmt.Sprintf("/collections/%s/points", collectionName))

	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to upsert points, status: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	return nil
}

// Search performs vector search in a collection
func (q *QdrantClient) Search(ctx context.Context, collectionName string, req SearchRequest) ([]SearchResult, error) {
	resp, err := q.client.R().
		SetContext(ctx).
		SetBody(req).
		Post(fmt.Sprintf("/collections/%s/points/search", collectionName))

	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("search failed, status: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	var result struct {
		Result []SearchResult `json:"result"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search result: %w", err)
	}

	return result.Result, nil
}

// DeleteCollection deletes a collection
func (q *QdrantClient) DeleteCollection(ctx context.Context, name string) error {
	resp, err := q.client.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/collections/%s", name))

	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete collection, status: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	return nil
}

// HealthCheck checks if Qdrant is healthy
func (q *QdrantClient) HealthCheck(ctx context.Context) error {
	resp, err := q.client.R().
		SetContext(ctx).
		Get("/health")

	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("health check failed, status: %d", resp.StatusCode())
	}

	return nil
}
