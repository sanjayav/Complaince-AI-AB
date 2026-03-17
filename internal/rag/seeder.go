package rag

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// SeederService populates Qdrant with sample data for testing
type SeederService struct {
	qdrantClient *QdrantClient
	embedder     *EmbedderService
}

// NewSeederService creates a new seeder service
func NewSeederService(qdrantURL string) *SeederService {
	return &SeederService{
		qdrantClient: NewQdrantClient(qdrantURL),
		embedder:     NewEmbedderService(),
	}
}

// SeedSampleData populates Qdrant with sample automotive test report data
func (s *SeederService) SeedSampleData(ctx context.Context) error {
	log.Println("Seeding Qdrant with sample data...")

	// Create collection if it doesn't exist
	collectionName := "document_entities"
	vectorSize := s.embedder.GetEmbeddingDimension()

	if err := s.qdrantClient.CreateCollection(ctx, collectionName, vectorSize); err != nil {
		log.Printf("Warning: Collection creation failed (may already exist): %v", err)
	}

	// Sample automotive test data
	sampleData := []struct {
		content string
		docType string
		docID   string
		page    int
		bbox    map[string]float64
	}{
		{
			content: "Engine performance test results show 0-60 mph acceleration in 4.2 seconds",
			docType: "chunk",
			docID:   "test-report-001",
			page:    1,
			bbox:    map[string]float64{"x": 100, "y": 150, "w": 400, "h": 30},
		},
		{
			content: "Brake system efficiency measured at 98.5% with ABS activation at 0.8g deceleration",
			docType: "chunk",
			docID:   "test-report-001",
			page:    2,
			bbox:    map[string]float64{"x": 120, "y": 200, "w": 450, "h": 35},
		},
		{
			content: "Fuel consumption: 28 mpg city, 35 mpg highway, 31 mpg combined",
			docType: "cell",
			docID:   "test-report-001",
			page:    3,
			bbox:    map[string]float64{"x": 200, "y": 300, "w": 300, "h": 25},
		},
		{
			content: "Safety rating: 5-star NCAP with advanced driver assistance systems",
			docType: "chunk",
			docID:   "test-report-002",
			page:    1,
			bbox:    map[string]float64{"x": 80, "y": 120, "w": 500, "h": 40},
		},
		{
			content: "Emissions compliance: Euro 6d-TEMP standard with NOx levels below 80 mg/km",
			docType: "chunk",
			docID:   "test-report-002",
			page:    4,
			bbox:    map[string]float64{"x": 150, "y": 250, "w": 420, "h": 30},
		},
		{
			content: "Vehicle weight distribution: 52% front, 48% rear for optimal handling",
			docType: "cell",
			docID:   "test-report-003",
			page:    2,
			bbox:    map[string]float64{"x": 180, "y": 180, "w": 350, "h": 28},
		},
		{
			content: "Tire performance: Dry grip rating 9.2/10, wet grip 8.8/10",
			docType: "chunk",
			docID:   "test-report-003",
			page:    5,
			bbox:    map[string]float64{"x": 90, "y": 320, "w": 480, "h": 35},
		},
		{
			content: "Battery capacity: 75 kWh with 280-mile range under WLTP conditions",
			docType: "cell",
			docID:   "test-report-004",
			page:    1,
			bbox:    map[string]float64{"x": 220, "y": 150, "w": 280, "h": 25},
		},
		{
			content: "Charging time: 0-80% in 30 minutes using 150 kW DC fast charger",
			docType: "chunk",
			docID:   "test-report-004",
			page:    3,
			bbox:    map[string]float64{"x": 110, "y": 280, "w": 460, "h": 30},
		},
		{
			content: "Noise levels: Interior cabin noise at 70 mph measured at 68 dB",
			docType: "chunk",
			docID:   "test-report-005",
			page:    2,
			bbox:    map[string]float64{"x": 160, "y": 220, "w": 380, "h": 32},
		},
	}

	// Generate embeddings for all content
	texts := make([]string, len(sampleData))
	for i, data := range sampleData {
		texts[i] = data.content
	}

	embeddings, err := s.embedder.Embed(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create points for Qdrant
	points := make([]Point, len(sampleData))
	for i, data := range sampleData {
		points[i] = Point{
			ID:     uuid.New().String(),
			Vector: embeddings[i],
			Payload: map[string]interface{}{
				"type":    data.docType,
				"doc_id":  data.docID,
				"page":    data.page,
				"content": data.content,
				"bbox":    data.bbox,
			},
		}
	}

	// Insert points into Qdrant
	if err := s.qdrantClient.UpsertPoints(ctx, collectionName, points); err != nil {
		return fmt.Errorf("failed to insert points: %w", err)
	}

	log.Printf("Successfully seeded %d sample entities into Qdrant collection '%s'", len(points), collectionName)
	return nil
}

// ClearSampleData removes all sample data from Qdrant
func (s *SeederService) ClearSampleData(ctx context.Context) error {
	log.Println("Clearing sample data from Qdrant...")

	collectionName := "document_entities"
	if err := s.qdrantClient.DeleteCollection(ctx, collectionName); err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	log.Println("Sample data cleared successfully")
	return nil
}
