package httpserver

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"jlrdi/internal/auth"
	"jlrdi/internal/rag"
	"jlrdi/internal/storage"

	"github.com/go-chi/chi/v5"
)

type API struct {
	deps Deps
}

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
	})
}

func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	writeJSON(w, http.StatusOK, user)
}

type semanticSearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
	Type  string `json:"type,omitempty"` // cell, figure, chunk, or empty for all
}

type bbox struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type evidence struct {
	Type    string  `json:"type"` // cell|figure|chunk
	ID      string  `json:"id"`
	DocID   string  `json:"doc_id"`
	Page    int     `json:"page"`
	BBox    bbox    `json:"bbox"`
	Score   float64 `json:"score"`
	Content string  `json:"content"`
}

type semanticSearchResponse struct {
	Query   string     `json:"query"`
	Results []evidence `json:"results"`
	Total   int        `json:"total"`
}

func (a *API) SemanticSearch(w http.ResponseWriter, r *http.Request) {
	var req semanticSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TopK <= 0 {
		req.TopK = 10
	}

	// Generate embedding for the query
	embedder := rag.NewEmbedderService()
	embeddings, err := embedder.Embed(r.Context(), []string{req.Query})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate embedding")
		return
	}

	// Search in Qdrant
	qdrantClient := rag.NewQdrantClient(a.deps.QdrantURL)
	searchReq := rag.SearchRequest{
		Vector:      embeddings[0],
		Limit:       req.TopK,
		WithPayload: true,
		WithVector:  false,
	}

	// Filter by type if specified
	if req.Type != "" {
		searchReq.Filter = map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"key":   "type",
					"match": map[string]interface{}{"value": req.Type},
				},
			},
		}
	}

	// Search in the appropriate collection
	collectionName := "document_entities"
	results, err := qdrantClient.Search(r.Context(), collectionName, searchReq)
	if err != nil {
		// If collection doesn't exist or search fails, return empty results
		writeJSON(w, http.StatusOK, semanticSearchResponse{
			Query:   req.Query,
			Results: []evidence{},
			Total:   0,
		})
		return
	}

	// Convert Qdrant results to evidence
	evidenceList := make([]evidence, 0, len(results))
	for _, result := range results {
		ev := evidence{
			Type:    getStringFromPayload(result.Payload, "type"),
			ID:      result.ID,
			DocID:   getStringFromPayload(result.Payload, "doc_id"),
			Page:    getIntFromPayload(result.Payload, "page"),
			Score:   result.Score,
			Content: getStringFromPayload(result.Payload, "content"),
		}

		// Parse bounding box if available
		if bboxData, ok := result.Payload["bbox"].(map[string]interface{}); ok {
			ev.BBox = bbox{
				X: getFloatFromMap(bboxData, "x"),
				Y: getFloatFromMap(bboxData, "y"),
				W: getFloatFromMap(bboxData, "w"),
				H: getFloatFromMap(bboxData, "h"),
			}
		}

		evidenceList = append(evidenceList, ev)
	}

	writeJSON(w, http.StatusOK, semanticSearchResponse{
		Query:   req.Query,
		Results: evidenceList,
		Total:   len(evidenceList),
	})
}

type askRequest struct {
	Question string `json:"question"`
	TopK     int    `json:"top_k"`
}

type citation struct {
	DocID        string `json:"doc_id"`
	Page         int    `json:"page"`
	BBox         bbox   `json:"bbox"`
	PDFURL       string `json:"pdf_url"`
	PageImageURL string `json:"page_image_url"`
	Content      string `json:"content"`
	Type         string `json:"type"`
}

type askResponse struct {
	Answer    string     `json:"answer"`
	Citations []citation `json:"citations"`
	Model     string     `json:"model"`
	Tokens    int        `json:"tokens"`
}

func (a *API) Ask(w http.ResponseWriter, r *http.Request) {
	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TopK <= 0 {
		req.TopK = 8
	}

	// First, perform semantic search to get relevant context
	embedder := rag.NewEmbedderService()
	embeddings, err := embedder.Embed(r.Context(), []string{req.Question})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate embedding")
		return
	}

	qdrantClient := rag.NewQdrantClient(a.deps.QdrantURL)
	searchReq := rag.SearchRequest{
		Vector:      embeddings[0],
		Limit:       req.TopK,
		WithPayload: true,
		WithVector:  false,
	}

	collectionName := "document_entities"
	results, err := qdrantClient.Search(r.Context(), collectionName, searchReq)
	if err != nil {
		// Return a generic response if search fails
		writeJSON(w, http.StatusOK, askResponse{
			Answer:    "I couldn't find relevant information to answer your question. Please try rephrasing or ask about a different topic.",
			Citations: []citation{},
			Model:     "stub-model",
			Tokens:    0,
		})
		return
	}

	// Convert results to citations
	citations := make([]citation, 0, len(results))
	for _, result := range results {
		citation := citation{
			DocID:   getStringFromPayload(result.Payload, "doc_id"),
			Page:    getIntFromPayload(result.Payload, "page"),
			Type:    getStringFromPayload(result.Payload, "type"),
			Content: getStringFromPayload(result.Payload, "content"),
		}

		// Parse bounding box
		if bboxData, ok := result.Payload["bbox"].(map[string]interface{}); ok {
			citation.BBox = bbox{
				X: getFloatFromMap(bboxData, "x"),
				Y: getFloatFromMap(bboxData, "y"),
				W: getFloatFromMap(bboxData, "w"),
				H: getFloatFromMap(bboxData, "h"),
			}
		}

		// Generate signed URLs for PDF and page image
		if citation.DocID != "" {
			citation.PDFURL, _ = a.presign(r.Context(), "docs/"+citation.DocID+".pdf", 15*time.Minute)
			citation.PageImageURL, _ = a.presign(r.Context(), fmt.Sprintf("pages/%s/%d.png", citation.DocID, citation.Page), 15*time.Minute)
		}

		citations = append(citations, citation)
	}

	// In a real implementation, this would call an LLM with the context
	// For now, generate a simple answer based on the citations
	answer := "Based on the available documents, I found some relevant information. "
	if len(citations) > 0 {
		answer += fmt.Sprintf("I found %d relevant pieces of evidence across %d documents.", len(citations), countUniqueDocs(citations))
	} else {
		answer += "However, I couldn't find specific evidence to answer your question."
	}

	writeJSON(w, http.StatusOK, askResponse{
		Answer:    answer,
		Citations: citations,
		Model:     "stub-model",
		Tokens:    len(req.Question) + len(answer),
	})
}

func (a *API) GetAnswer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// In a real implementation, fetch from database
	writeJSON(w, http.StatusOK, map[string]any{
		"id":         id,
		"answer":     "This is a stub answer. In production, fetch from database.",
		"question":   "What was the original question?",
		"created_at": time.Now().UTC().Format(time.RFC3339),
	})
}

type manifestPage struct {
	Page     int    `json:"page"`
	ImageURL string `json:"image_url"`
}

type documentManifest struct {
	DocID  string         `json:"doc_id"`
	PDFURL string         `json:"pdf_url"`
	Pages  []manifestPage `json:"pages"`
	Status string         `json:"status"`
}

func (a *API) DocumentManifest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := chi.URLParam(r, "docId")

	// In a real implementation, fetch document metadata from database
	pdfKey := "docs/" + docID + ".pdf"
	pdfURL, _ := a.presign(ctx, pdfKey, 15*time.Minute)

	// Provide stub pages (in production, fetch from database)
	pages := make([]manifestPage, 0, 3)
	for i := 1; i <= 3; i++ {
		key := "pages/" + docID + "/" + strconv.Itoa(i) + ".png"
		url, _ := a.presign(ctx, key, 15*time.Minute)
		pages = append(pages, manifestPage{Page: i, ImageURL: url})
	}

	writeJSON(w, http.StatusOK, documentManifest{
		DocID:  docID,
		PDFURL: pdfURL,
		Pages:  pages,
		Status: "processed",
	})
}

type highlightSignedURLsRequest struct {
	DocID   string `json:"doc_id"`
	Page    int    `json:"page"`
	Regions []bbox `json:"regions"`
}

type highlightSignedURLsResponse struct {
	PDFURL       string   `json:"pdf_url"`
	PageImageURL string   `json:"page_image_url"`
	RegionURLs   []string `json:"region_urls"`
}

func (a *API) HighlightSignedURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req highlightSignedURLsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	pdfURL, _ := a.presign(ctx, "docs/"+req.DocID+".pdf", 15*time.Minute)
	pageURL, _ := a.presign(ctx, "pages/"+req.DocID+"/"+strconv.Itoa(req.Page)+".png", 15*time.Minute)

	regionURLs := make([]string, len(req.Regions))
	for i := range req.Regions {
		// In a real system, region crops would be generated and stored
		regionURLs[i] = pageURL
	}

	writeJSON(w, http.StatusOK, highlightSignedURLsResponse{
		PDFURL:       pdfURL,
		PageImageURL: pageURL,
		RegionURLs:   regionURLs,
	})
}

func (a *API) GetEvidence(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	id := chi.URLParam(r, "id")

	// In a real implementation, fetch from database
	writeJSON(w, http.StatusOK, map[string]any{
		"type": typ,
		"id":   id,
		"meta": map[string]any{
			"stub":      true,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (a *API) Export(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=export.csv")
	w.WriteHeader(http.StatusOK)

	csvw := csv.NewWriter(w)
	_ = csvw.Write([]string{"doc_id", "page", "type", "id", "text", "bbox_x", "bbox_y", "bbox_w", "bbox_h"})
	_ = csvw.Write([]string{"example-doc", "1", "chunk", "chunk-1", "example text content", "100.0", "200.0", "300.0", "50.0"})
	csvw.Flush()
}

func (a *API) IndexEnqueue(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would add a job to a queue
	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":    "enqueued",
		"job_id":    "job-" + strconv.FormatInt(time.Now().Unix(), 10),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *API) QAPendingTasks(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, fetch from database
	writeJSON(w, http.StatusOK, []map[string]any{
		{
			"id":         "task-1",
			"type":       "classification",
			"status":     "pending",
			"created_at": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (a *API) QAApprove(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	writeJSON(w, http.StatusOK, map[string]any{
		"id":         id,
		"status":     "approved",
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *API) QAReject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	writeJSON(w, http.StatusOK, map[string]any{
		"id":         id,
		"status":     "rejected",
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// SeedData populates Qdrant with sample data for testing
func (a *API) SeedData(w http.ResponseWriter, r *http.Request) {
	seeder := rag.NewSeederService(a.deps.QdrantURL)

	if err := seeder.SeedSampleData(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to seed data: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "seeded",
		"message":   "Sample data successfully seeded into Qdrant",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ClearData removes all sample data from Qdrant
func (a *API) ClearData(w http.ResponseWriter, r *http.Request) {
	seeder := rag.NewSeederService(a.deps.QdrantURL)

	if err := seeder.ClearSampleData(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to clear data: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "cleared",
		"message":   "Sample data successfully cleared from Qdrant",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// UploadDocument handles PDF document uploads
func (a *API) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// Support both multipart/form-data uploads and raw PDF stream uploads
	var (
		filename    string
		contentType string
		source      io.ReadCloser
		docType     string
		sizeHint    int64
	)

	contentTypeHeader := r.Header.Get("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentTypeHeader), "multipart/form-data") {
		// Multipart path
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
			writeError(w, http.StatusBadRequest, "failed to parse form")
			return
		}

		f, hdr, err := r.FormFile("document")
		if err != nil {
			f, hdr, err = r.FormFile("file")
			if err != nil {
				writeError(w, http.StatusBadRequest, "no file provided")
				return
			}
		}
		source = f
		filename = hdr.Filename
		contentType = hdr.Header.Get("Content-Type")
		sizeHint = hdr.Size
		docType = r.FormValue("type")
		if docType == "" {
			docType = "automotive_test"
		}
	} else {
		// Raw upload path (e.g., application/pdf or application/octet-stream)
		source = r.Body
		filename = r.URL.Query().Get("filename")
		if filename == "" {
			filename = r.Header.Get("X-Filename")
		}
		if filename == "" {
			filename = "upload.pdf"
		}
		contentType = contentTypeHeader
		if contentType == "" {
			contentType = "application/pdf"
		}
		docType = r.URL.Query().Get("type")
		if docType == "" {
			docType = "automotive_test"
		}
	}
	defer func() {
		if source != nil {
			_ = source.Close()
		}
	}()

	// Basic validation: ensure a PDF filename
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		writeError(w, http.StatusBadRequest, "only PDF files are supported")
		return
	}

	// Create S3 service
	s3Service, err := storage.NewS3Service(a.deps.Config.S3Bucket, "us-east-1")
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create S3 service: %v", err))
		return
	}
	defer s3Service.Close()

	// Create a temporary file for processing
	tempFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp file")
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy uploaded content to temp file
	bytesWritten, err := io.Copy(tempFile, source)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save temp file")
		return
	}

	// Reset file pointer to beginning for S3 upload
	if _, err := tempFile.Seek(0, 0); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset temp file")
		return
	}

	// Determine size to use for S3 (prefer hint when available)
	sizeForS3 := sizeHint
	if sizeForS3 <= 0 {
		sizeForS3 = bytesWritten
	}

	// Upload to S3 using the file reader
	docInfo, err := s3Service.UploadDocumentFromReader(r.Context(), filename, tempFile, sizeForS3, contentType, docType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to upload to S3: %v", err))
		return
	}

	// Trigger document processing (async)
	go func() {
		if err := a.processDocumentAsync(r.Context(), tempFile.Name(), docInfo.Key, docType); err != nil {
			log.Printf("Error processing document %s: %v", docInfo.Key, err)
		}
	}()

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":      "uploaded",
		"document_id": strings.TrimSuffix(filepath.Base(docInfo.Key), filepath.Ext(docInfo.Key)),
		"s3_key":      docInfo.Key,
		"size":        docInfo.Size,
		"message":     "Document uploaded successfully and processing started",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	})
}

// processDocumentAsync processes the uploaded document asynchronously
func (a *API) processDocumentAsync(ctx context.Context, tempFilePath, s3Key, docType string) error {
	log.Printf("Starting async processing of document: %s", s3Key)

	// Extract document ID from S3 key
	docID := strings.TrimSuffix(filepath.Base(s3Key), filepath.Ext(s3Key))

	// Create PDF processor
	pdfProcessor, err := rag.NewPDFProcessor("/tmp/jlrdi/pages")
	if err != nil {
		return fmt.Errorf("failed to create PDF processor: %w", err)
	}
	defer pdfProcessor.Close()

	// Process the PDF
	processedPages, err := pdfProcessor.ProcessPDF(ctx, tempFilePath, docID)
	if err != nil {
		return fmt.Errorf("failed to process PDF: %w", err)
	}

	log.Printf("Successfully processed %d pages from document %s", len(processedPages), docID)

	// Extract content for embedding and storage
	var allContent []string
	var allTables []rag.Table
	var allFigures []rag.Figure

	for _, page := range processedPages {
		// Add page text
		if page.Text != "" {
			allContent = append(allContent, page.Text)
		}

		// Add tables
		for _, table := range page.Tables {
			allTables = append(allTables, table)
			// Create table content string for embedding
			var tableContent strings.Builder
			for _, row := range table.Rows {
				tableContent.WriteString(strings.Join(row, " | "))
				tableContent.WriteString("\n")
			}
			allContent = append(allContent, tableContent.String())
		}

		// Add figures
		for _, figure := range page.Figures {
			allFigures = append(allFigures, figure)
			if figure.Caption != "" {
				allContent = append(allContent, figure.Caption)
			}
		}
	}

	// Generate embeddings for extracted content
	if len(allContent) > 0 {
		embedder := rag.NewEmbedderService()
		embeddings, err := embedder.Embed(ctx, allContent)
		if err != nil {
			log.Printf("Warning: failed to generate embeddings: %v", err)
		} else {
			log.Printf("Generated %d embeddings for document %s", len(embeddings), docID)

			// Store in Qdrant (this would be enhanced in production)
			// For now, just log the success
			log.Printf("Content from document %s ready for vector storage", docID)
		}
	}

	log.Printf("Document processing completed: %d pages, %d tables, %d figures",
		len(processedPages), len(allTables), len(allFigures))

	return nil
}

// ListDocuments lists uploaded documents
func (a *API) ListDocuments(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = "documents/"
	}

	maxKeys := int32(100)
	if maxStr := r.URL.Query().Get("max_keys"); maxStr != "" {
		if max, err := strconv.ParseInt(maxStr, 10, 32); err == nil {
			maxKeys = int32(max)
		}
	}

	// Create S3 service
	s3Service, err := storage.NewS3Service(a.deps.Config.S3Bucket, "us-east-1")
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create S3 service: %v", err))
		return
	}
	defer s3Service.Close()

	// List documents
	documents, err := s3Service.ListDocuments(r.Context(), prefix, maxKeys)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list documents: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"documents": documents,
		"total":     len(documents),
		"prefix":    prefix,
	})
}

func (a *API) presign(ctx context.Context, key string, expires time.Duration) (string, error) {
	return a.deps.Signer.PresignGetObject(ctx, a.deps.Config.S3Bucket, key, expires)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

// Helper functions for payload parsing
func getStringFromPayload(payload map[string]interface{}, key string) string {
	if val, ok := payload[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntFromPayload(payload map[string]interface{}, key string) int {
	if val, ok := payload[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
		if num, ok := val.(int); ok {
			return num
		}
	}
	return 0
}

func getFloatFromMap(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if num, ok := val.(float64); ok {
			return num
		}
	}
	return 0.0
}

func countUniqueDocs(citations []citation) int {
	seen := make(map[string]bool)
	count := 0
	for _, c := range citations {
		if c.DocID != "" && !seen[c.DocID] {
			seen[c.DocID] = true
			count++
		}
	}
	return count
}
