package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"jlrdi/internal/auth"
	"jlrdi/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentAPI handles document management operations
type DocumentAPI struct {
	tenantS3Service *storage.TenantS3Service
	db              *pgxpool.Pool
}

// NewDocumentAPI creates a new document API handler
func NewDocumentAPI(tenantS3Service *storage.TenantS3Service, db *pgxpool.Pool) *DocumentAPI {
	return &DocumentAPI{
		tenantS3Service: tenantS3Service,
		db:              db,
	}
}

// UploadDocumentRequest represents a request to upload a document
type UploadDocumentRequest struct {
	TenantID string            `json:"tenant_id"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UploadDocumentResponse represents the response when uploading a document
type UploadDocumentResponse struct {
	Document *storage.TenantDocument `json:"document"`
	Message  string                  `json:"message"`
}

// UploadDocument uploads a single document for a tenant
func (d *DocumentAPI) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to parse form"})
		return
	}

	// Get tenant ID from form or header
	tenantID := r.FormValue("tenant_id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate constraints: PDF only, max 5 MB
	const maxSize = 5 * 1024 * 1024
	contentType := header.Header.Get("Content-Type")
	if header.Size > maxSize {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "File too large. Max size is 5MB"})
		return
	}
	if contentType != "application/pdf" && !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Only PDF files are allowed"})
		return
	}

	// Parse metadata
	metadata := make(map[string]string)
	if metadataStr := r.FormValue("metadata"); metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid metadata format"})
			return
		}
	}

	// Add file info to metadata
	metadata["original_filename"] = header.Filename
	metadata["content_type"] = header.Header.Get("Content-Type")
	metadata["upload_timestamp"] = time.Now().Format(time.RFC3339)

	// Upload document
	document, err := d.tenantS3Service.UploadDocument(
		r.Context(),
		tenantID,
		file,
		header.Filename,
		header.Header.Get("Content-Type"),
		metadata,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Upload failed: %v", err)})
		return
	}

	response := UploadDocumentResponse{
		Document: document,
		Message:  "Document uploaded successfully",
	}

	writeJSON(w, http.StatusCreated, response)
}

// UploadMyDocument uploads a single document for the authenticated user's tenant (no tenant_id param)
func (d *DocumentAPI) UploadMyDocument(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user.Subject == "" || user.TenantID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to parse form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate constraints: PDF only, max 5 MB
	const maxSize = 5 * 1024 * 1024
	contentType := header.Header.Get("Content-Type")
	if header.Size > maxSize {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "File too large. Max size is 5MB"})
		return
	}
	if contentType != "application/pdf" && !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Only PDF files are allowed"})
		return
	}

	metadata := map[string]string{
		"original_filename": header.Filename,
		"content_type":      header.Header.Get("Content-Type"),
		"upload_timestamp":  time.Now().Format(time.RFC3339),
	}

	doc, err := d.tenantS3Service.UploadDocument(r.Context(), user.TenantID, file, header.Filename, header.Header.Get("Content-Type"), metadata)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Upload failed: %v", err)})
		return
	}

	// Persist a row in DB (best-effort)
	if d.db != nil {
		s3key := fmt.Sprintf("tenants/%s/documents/%s", user.TenantID, doc.ID)
		s3url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", "jlrdi", "us-east-1", s3key)
		_, _ = d.db.Exec(r.Context(), `INSERT INTO documents (filename, s3_key, file_size, mime_type, status, metadata, tenant_id, s3_url) VALUES ($1,$2,$3,$4,'uploaded',$5,$6,$7)`, doc.Filename, s3key, doc.Size, doc.ContentType, metadata, user.TenantID, s3url)
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":       doc.ID,
		"filename": doc.Filename,
		"size":     doc.Size,
		"message":  "uploaded",
	})
}

// BatchUploadRequest represents a request to upload multiple documents
type BatchUploadRequest struct {
	TenantID      string            `json:"tenant_id"`
	MaxConcurrent int               `json:"max_concurrent,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// BatchUploadResponse represents the response when batch uploading documents
type BatchUploadResponse struct {
	Result  *storage.BatchUploadResult `json:"result"`
	Message string                     `json:"message"`
}

// BatchUploadDocuments uploads multiple documents concurrently
func (d *DocumentAPI) BatchUploadDocuments(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to parse form"})
		return
	}

	// Get tenant ID from form
	tenantID := r.FormValue("tenant_id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// Get max concurrent from form
	maxConcurrent := 5 // Default
	if maxConcurrentStr := r.FormValue("max_concurrent"); maxConcurrentStr != "" {
		if val, err := strconv.Atoi(maxConcurrentStr); err == nil && val > 0 {
			maxConcurrent = val
		}
	}

	// Parse metadata
	metadata := make(map[string]string)
	if metadataStr := r.FormValue("metadata"); metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid metadata format"})
			return
		}
	}

	// Get files from form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No files provided"})
		return
	}

	// Convert to BatchFile format
	batchFiles := make([]storage.BatchFile, 0, len(files))
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to open file %s: %v", fileHeader.Filename, err)})
			return
		}

		// Create a copy of metadata for this file
		fileMetadata := make(map[string]string)
		for k, v := range metadata {
			fileMetadata[k] = v
		}
		fileMetadata["original_filename"] = fileHeader.Filename
		fileMetadata["content_type"] = fileHeader.Header.Get("Content-Type")
		fileMetadata["upload_timestamp"] = time.Now().Format(time.RFC3339)

		batchFiles = append(batchFiles, storage.BatchFile{
			Reader:      file,
			Filename:    fileHeader.Filename,
			ContentType: fileHeader.Header.Get("Content-Type"),
			Size:        fileHeader.Size,
			Metadata:    fileMetadata,
		})
	}

	// Perform batch upload
	result := d.tenantS3Service.BatchUploadDocuments(r.Context(), tenantID, batchFiles, maxConcurrent)

	response := BatchUploadResponse{
		Result:  result,
		Message: fmt.Sprintf("Batch upload completed: %d success, %d failures", result.SuccessCount, result.FailureCount),
	}

	writeJSON(w, http.StatusOK, response)
}

// MyDocuments lists documents for the authenticated user's tenant with filename, size and presigned URL
func (d *DocumentAPI) ListMyDocuments(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user.Subject == "" || user.TenantID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	documents, err := d.tenantS3Service.ListTenantDocuments(r.Context(), user.TenantID, "", 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to list documents: %v", err)})
		return
	}

	type item struct {
		ID       string `json:"id"`
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		URL      string `json:"url"`
	}
	list := make([]item, 0, len(documents))
	for _, doc := range documents {
		url, _ := d.tenantS3Service.GetDocumentPresignedURL(r.Context(), user.TenantID, doc.ID, 3600*time.Second)
		list = append(list, item{ID: doc.ID, Filename: doc.Filename, Size: doc.Size, URL: url})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"files": list, "total": len(list)})
}

// DeleteMyDocument deletes a document by ID for the authenticated user's tenant
func (d *DocumentAPI) DeleteMyDocument(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user.Subject == "" || user.TenantID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	did := chi.URLParam(r, "did")
	if did == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "document id required"})
		return
	}
	if err := d.tenantS3Service.DeleteDocument(r.Context(), user.TenantID, did); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete document: %v", err)})
		return
	}
	// Best-effort DB cleanup
	if d.db != nil {
		_, _ = d.db.Exec(r.Context(), `DELETE FROM documents WHERE tenant_id=$1 AND s3_key LIKE 'tenants/'||$1||'/documents/'||$2||'%'`, user.TenantID, did)
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ListDocumentsRequest represents a request to list documents
type ListDocumentsRequest struct {
	TenantID string `json:"tenant_id"`
	Prefix   string `json:"prefix,omitempty"`
	MaxKeys  int32  `json:"max_keys,omitempty"`
}

// ListDocumentsResponse represents the response when listing documents
type ListDocumentsResponse struct {
	Documents []storage.TenantDocument `json:"documents"`
	Total     int                      `json:"total"`
	Message   string                   `json:"message"`
}

// ListDocuments lists documents for a tenant
func (d *DocumentAPI) ListDocuments(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from query params
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// Get optional parameters
	prefix := r.URL.Query().Get("prefix")
	maxKeys := int32(100) // Default
	if maxKeysStr := r.URL.Query().Get("max_keys"); maxKeysStr != "" {
		if val, err := strconv.ParseInt(maxKeysStr, 10, 32); err == nil && val > 0 {
			maxKeys = int32(val)
		}
	}

	// List documents
	documents, err := d.tenantS3Service.ListTenantDocuments(r.Context(), tenantID, prefix, maxKeys)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to list documents: %v", err)})
		return
	}

	response := ListDocumentsResponse{
		Documents: documents,
		Total:     len(documents),
		Message:   "Documents retrieved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

// GetDocument retrieves a document by ID
func (d *DocumentAPI) GetDocument(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	documentID := chi.URLParam(r, "did")

	if tenantID == "" || documentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID and Document ID are required"})
		return
	}

	// Get document
	document, err := d.tenantS3Service.GetDocument(r.Context(), tenantID, documentID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Document not found: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"document": document,
		"message":  "Document retrieved successfully",
	})
}

// DeleteDocument deletes a document
func (d *DocumentAPI) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	documentID := chi.URLParam(r, "did")

	if tenantID == "" || documentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID and Document ID are required"})
		return
	}

	// Delete document
	err := d.tenantS3Service.DeleteDocument(r.Context(), tenantID, documentID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to delete document: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Document deleted successfully",
	})
}

// GetDocumentPresignedURL generates a presigned URL for document access
func (d *DocumentAPI) GetDocumentPresignedURL(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	documentID := chi.URLParam(r, "did")

	if tenantID == "" || documentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID and Document ID are required"})
		return
	}

	// Get expiration from query params
	expiration := 1 * time.Hour // Default
	if expirationStr := r.URL.Query().Get("expiration"); expirationStr != "" {
		if val, err := strconv.ParseInt(expirationStr, 10, 64); err == nil && val > 0 {
			expiration = time.Duration(val) * time.Second
		}
	}

	// Generate presigned URL
	url, err := d.tenantS3Service.GetDocumentPresignedURL(r.Context(), tenantID, documentID, expiration)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to generate presigned URL: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"presigned_url": url,
		"expiration":    expiration.Seconds(),
		"message":       "Presigned URL generated successfully",
	})
}

// UpdateDocumentStatusRequest represents a request to update document status
type UpdateDocumentStatusRequest struct {
	Status   string            `json:"status"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UpdateDocumentStatus updates the processing status of a document
func (d *DocumentAPI) UpdateDocumentStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	documentID := chi.URLParam(r, "did")

	if tenantID == "" || documentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID and Document ID are required"})
		return
	}

	var req UpdateDocumentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate status
	validStatuses := map[string]bool{"uploaded": true, "processing": true, "processed": true, "failed": true}
	if !validStatuses[req.Status] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid status"})
		return
	}

	// Update status
	err := d.tenantS3Service.UpdateDocumentStatus(r.Context(), tenantID, documentID, req.Status, req.Metadata)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to update status: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Document status updated successfully",
	})
}

// GetTenantStorageStats returns storage statistics for a tenant
func (d *DocumentAPI) GetTenantStorageStats(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Tenant ID is required"})
		return
	}

	// Get storage stats
	stats, err := d.tenantS3Service.GetTenantStorageStats(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to get storage stats: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"stats":   stats,
		"message": "Storage statistics retrieved successfully",
	})
}
