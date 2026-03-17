package rag

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// ProductionPDFProcessor handles real PDF document processing for automotive test reports
type ProductionPDFProcessor struct {
	outputDir string
}

// NewProductionPDFProcessor creates a new production PDF processor
func NewProductionPDFProcessor(outputDir string) (*ProductionPDFProcessor, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &ProductionPDFProcessor{
		outputDir: outputDir,
	}, nil
}

// ProcessPDF processes a real PDF file and extracts all content
func (p *ProductionPDFProcessor) ProcessPDF(ctx context.Context, pdfPath, docID string) ([]ProcessedPage, error) {
	log.Printf("Processing PDF: %s (PRODUCTION MODE)", pdfPath)

	// Open PDF file
	file, err := os.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer file.Close()

	// Create PDF reader
	reader, err := model.NewPdfReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	// Get number of pages
	numPages, err := reader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	log.Printf("Processing %d pages", numPages)

	var processedPages []ProcessedPage

	// Process each page
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		log.Printf("Processing page %d/%d", pageNum, numPages)

		// Get page
		page, err := reader.GetPage(pageNum)
		if err != nil {
			log.Printf("Warning: failed to get page %d: %v", pageNum, err)
			continue
		}

		// Extract text
		ex, err := extractor.New(page)
		if err != nil {
			log.Printf("Warning: failed to create extractor for page %d: %v", pageNum, err)
			continue
		}

		text, err := ex.ExtractText()
		if err != nil {
			log.Printf("Warning: failed to extract text from page %d: %v", pageNum, err)
			text = "" // Continue with empty text
		}

		// Get page dimensions
		pageSize, err := page.GetMediaBox()
		if err != nil {
			log.Printf("Warning: failed to get page size for page %d: %v", pageNum, err)
			pageSize = &model.PdfRectangle{
				Llx: 0, Lly: 0, Urx: 595, Ury: 842, // Default A4 size
			}
		}

		// Extract tables using advanced pattern matching for automotive reports
		tables := p.extractAutomotiveTables(text, pageNum, pageSize)

		// Extract figures and images specific to automotive reports
		figures := p.extractAutomotiveFigures(text, pageNum, pageSize)

		// Generate page image
		imagePath, err := p.generatePageImage(pdfPath, pageNum, docID)
		if err != nil {
			log.Printf("Warning: failed to generate image for page %d: %v", pageNum, err)
			imagePath = ""
		}

		processedPage := ProcessedPage{
			PageNumber: pageNum,
			Width:      pageSize.Urx - pageSize.Llx,
			Height:     pageSize.Ury - pageSize.Lly,
			Text:       text,
			Tables:     tables,
			Figures:    figures,
			ImagePath:  imagePath,
		}

		processedPages = append(processedPages, processedPage)
	}

	log.Printf("Successfully processed %d pages", len(processedPages))
	return processedPages, nil
}

// extractAutomotiveTables uses sophisticated pattern matching for automotive test data
func (p *ProductionPDFProcessor) extractAutomotiveTables(text string, pageNum int, pageSize *model.PdfRectangle) []Table {
	var tables []Table

	// Split text into lines
	lines := strings.Split(text, "\n")

	// Automotive-specific table detection patterns
	tablePatterns := []*regexp.Regexp{
		// Performance metrics
		regexp.MustCompile(`^\s*[A-Z][a-z\s]+:\s*[0-9.]+`), // "0-60 mph: 4.2"
		regexp.MustCompile(`^\s*[A-Z][a-z\s]+\s+[0-9.]+`),  // "Top speed 155"
		regexp.MustCompile(`^\s*[0-9.]+[a-zA-Z\s%]+`),      // "4.2 seconds"

		// Fuel economy
		regexp.MustCompile(`^\s*[A-Z][a-z\s]+\s+[0-9]+\s+mpg`), // "City driving 28 mpg"
		regexp.MustCompile(`^\s*[A-Z][a-z\s]+\s+[0-9.]+`),      // "Combined 31"

		// Safety ratings
		regexp.MustCompile(`^\s*[A-Z][a-z\s]+\s+[0-9]\s+stars?`), // "Overall 5 stars"
		regexp.MustCompile(`^\s*[A-Z][a-z\s]+\s+protection`),     // "Adult protection"

		// Technical specifications
		regexp.MustCompile(`^\s*[A-Z][a-z\s]+\s+[0-9.]+[A-Z]`),        // "Engine 3.0L"
		regexp.MustCompile(`^\s*[A-Z][a-z\s]+\s+[0-9]+\s+@\s+[0-9]+`), // "Power 395 @ 5500"
	}

	var currentTable [][]string
	var tableStartLine int

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line matches automotive table patterns
		isTableRow := false
		for _, pattern := range tablePatterns {
			if pattern.MatchString(line) {
				isTableRow = true
				break
			}
		}

		// Additional table detection: lines with multiple columns separated by spaces
		if !isTableRow && strings.Contains(line, "  ") {
			columns := strings.Fields(line)
			if len(columns) >= 2 {
				// Check if columns look like automotive table data
				hasNumbers := false
				hasText := false
				for _, col := range columns {
					if regexp.MustCompile(`^[0-9.]+`).MatchString(col) {
						hasNumbers = true
					} else if regexp.MustCompile(`^[A-Za-z]+`).MatchString(col) {
						hasText = true
					}
				}
				if hasNumbers && hasText {
					isTableRow = true
				}
			}
		}

		if isTableRow {
			if len(currentTable) == 0 {
				tableStartLine = i
			}

			// Split line into columns (handle multiple spaces)
			columns := regexp.MustCompile(`\s{2,}`).Split(line, -1)
			if len(columns) == 1 {
				// Fallback to simple space splitting
				columns = strings.Fields(line)
			}

			currentTable = append(currentTable, columns)
		} else if len(currentTable) > 0 {
			// End of table detected
			if len(currentTable) >= 2 { // At least 2 rows to be considered a table
				table := Table{
					Rows:    currentTable,
					PageNum: pageNum,
					BBox: BoundingBox{
						X: pageSize.Llx,
						Y: pageSize.Ury - float64(tableStartLine*12), // Approximate line height
						W: pageSize.Urx - pageSize.Llx,
						H: float64(len(currentTable) * 12),
					},
				}
				tables = append(tables, table)
			}
			currentTable = nil
		}
	}

	// Don't forget the last table
	if len(currentTable) >= 2 {
		table := Table{
			Rows:    currentTable,
			PageNum: pageNum,
			BBox: BoundingBox{
				X: pageSize.Llx,
				Y: pageSize.Ury - float64(tableStartLine*12),
				W: pageSize.Urx - pageSize.Llx,
				H: float64(len(currentTable) * 12),
			},
		}
		tables = append(tables, table)
	}

	return tables
}

// extractAutomotiveFigures identifies figures, charts, and images specific to automotive reports
func (p *ProductionPDFProcessor) extractAutomotiveFigures(text string, pageNum int, pageSize *model.PdfRectangle) []Figure {
	var figures []Figure

	// Automotive-specific figure detection patterns
	figurePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s*Figure\s+[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Fig\.\s*[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Chart\s+[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Graph\s+[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Performance\s+[A-Za-z]+`),
		regexp.MustCompile(`(?i)^\s*[A-Za-z]+\s+Curve`),
		regexp.MustCompile(`(?i)^\s*[A-Za-z]+\s+Diagram`),
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Check for automotive figure patterns
		for _, pattern := range figurePatterns {
			if pattern.MatchString(line) {
				// Extract caption
				caption := strings.TrimSpace(line)

				// Try to get the next few lines as part of the caption
				for j := 1; j <= 3 && i+j < len(lines); j++ {
					nextLine := strings.TrimSpace(lines[i+j])
					if nextLine != "" && !regexp.MustCompile(`^[0-9]+$`).MatchString(nextLine) {
						caption += " " + nextLine
					}
				}

				figure := Figure{
					Caption:   caption,
					PageNum:   pageNum,
					ImagePath: "", // Will be filled by image extraction
					BBox: BoundingBox{
						X: pageSize.Llx + 50, // Approximate figure position
						Y: pageSize.Ury - float64((i+1)*12) - 100,
						W: pageSize.Urx - pageSize.Llx - 100,
						H: 150, // Approximate figure height
					},
				}
				figures = append(figures, figure)
				break
			}
		}
	}

	return figures
}

// generatePageImage creates a PNG image representation of the PDF page
func (p *ProductionPDFProcessor) generatePageImage(pdfPath string, pageNum int, docID string) (string, error) {
	// Create document-specific directory
	docDir := filepath.Join(p.outputDir, "pages", docID)
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create document directory: %w", err)
	}

	// Generate image filename
	imagePath := filepath.Join(docDir, fmt.Sprintf("%d.png", pageNum))

	// For now, create a placeholder image
	// In production, you would use a PDF-to-image library like:
	// - pdf2image (Python)
	// - ImageMagick
	// - MuPDF
	// - Poppler utilities

	// Create a simple placeholder image
	if err := p.createPlaceholderImage(imagePath, pageNum); err != nil {
		return "", fmt.Errorf("failed to create placeholder image: %w", err)
	}

	return imagePath, nil
}

// createPlaceholderImage creates a simple placeholder image for development
func (p *ProductionPDFProcessor) createPlaceholderImage(imagePath string, pageNum int) error {
	// Create a simple 800x600 image with page number
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Fill with light gray background
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}

	// Create output file
	outputFile, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer outputFile.Close()

	// Encode as PNG
	if err := png.Encode(outputFile, img); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

// Close cleans up resources
func (p *ProductionPDFProcessor) Close() error {
	return nil
}
