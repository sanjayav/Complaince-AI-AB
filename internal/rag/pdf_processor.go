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
)

// PDFProcessor handles PDF document processing and extraction
type PDFProcessor struct {
	outputDir string
}

// ProcessedPage represents a processed page with extracted content
type ProcessedPage struct {
	PageNumber int
	Width      float64
	Height     float64
	Text       string
	Tables     []Table
	Figures    []Figure
	ImagePath  string
}

// Table represents an extracted table
type Table struct {
	Rows    [][]string
	BBox    BoundingBox
	PageNum int
}

// Figure represents an extracted figure or image
type Figure struct {
	Caption   string
	BBox      BoundingBox
	PageNum   int
	ImagePath string
}

// BoundingBox represents coordinates and dimensions
type BoundingBox struct {
	X, Y, W, H float64
}

// NewPDFProcessor creates a new PDF processor
func NewPDFProcessor(outputDir string) (*PDFProcessor, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &PDFProcessor{
		outputDir: outputDir,
	}, nil
}

// ProcessPDF processes a PDF file and extracts all content
// This is a development version that simulates PDF processing
// In production, this would use actual PDF libraries like UniPDF, Apache PDFBox, or PyMuPDF
func (p *PDFProcessor) ProcessPDF(ctx context.Context, pdfPath, docID string) ([]ProcessedPage, error) {
	log.Printf("Processing PDF: %s (development mode)", pdfPath)

	// For development, we'll simulate PDF processing based on file content
	// In production, this would:
	// 1. Open and parse the actual PDF
	// 2. Extract text using PDF libraries
	// 3. Detect tables using layout analysis
	// 4. Identify figures and images
	// 5. Generate page images

	// Simulate processing 2 pages for now
	processedPages := []ProcessedPage{
		{
			PageNumber: 1,
			Width:      595.0, // A4 width in points
			Height:     842.0, // A4 height in points
			Text:       "JLR AUTOMOTIVE TEST REPORT\nTest Report ID: TR-2024-001\nDate: 2024-01-15\nVehicle: Range Rover Sport\n\nENGINE PERFORMANCE TEST RESULTS\n0-60 mph acceleration: 4.2 seconds\nTop speed: 155 mph (electronically limited)\nEngine displacement: 3.0L I6\nPower output: 395 hp @ 5,500 rpm\nTorque: 406 lb-ft @ 2,000 rpm",
			Tables: []Table{
				{
					Rows: [][]string{
						{"Metric", "Value", "Unit"},
						{"0-60 mph", "4.2", "seconds"},
						{"Top speed", "155", "mph"},
						{"Power output", "395", "hp"},
						{"Torque", "406", "lb-ft"},
					},
					PageNum: 1,
					BBox:    BoundingBox{X: 100, Y: 200, W: 400, H: 120},
				},
			},
			Figures: []Figure{
				{
					Caption:   "Figure 1: Performance Chart",
					PageNum:   1,
					ImagePath: fmt.Sprintf("pages/%s/1.png", docID),
					BBox:      BoundingBox{X: 300, Y: 400, W: 200, H: 150},
				},
			},
			ImagePath: fmt.Sprintf("pages/%s/1.png", docID),
		},
		{
			PageNumber: 2,
			Width:      595.0,
			Height:     842.0,
			Text:       "FUEL CONSUMPTION\nCity driving: 28 mpg\nHighway driving: 35 mpg\nCombined: 31 mpg\nFuel tank capacity: 23.8 gallons\n\nSAFETY RATINGS\nNCAP overall rating: 5 stars\nAdult occupant protection: 5 stars\nChild occupant protection: 5 stars\nVulnerable road user protection: 4 stars\nSafety assist systems: 5 stars",
			Tables: []Table{
				{
					Rows: [][]string{
						{"Driving Mode", "MPG"},
						{"City", "28"},
						{"Highway", "35"},
						{"Combined", "31"},
					},
					PageNum: 2,
					BBox:    BoundingBox{X: 100, Y: 150, W: 300, H: 100},
				},
				{
					Rows: [][]string{
						{"Safety Category", "Rating"},
						{"Overall", "5 stars"},
						{"Adult Protection", "5 stars"},
						{"Child Protection", "5 stars"},
						{"Road User Protection", "4 stars"},
						{"Safety Assist", "5 stars"},
					},
					PageNum: 2,
					BBox:    BoundingBox{X: 100, Y: 350, W: 350, H: 150},
				},
			},
			Figures:   []Figure{},
			ImagePath: fmt.Sprintf("pages/%s/2.png", docID),
		},
	}

	// Generate page images
	for i := range processedPages {
		imagePath, err := p.generatePageImage(processedPages[i].ImagePath, processedPages[i].PageNumber, docID)
		if err != nil {
			log.Printf("Warning: failed to generate image for page %d: %v", processedPages[i].PageNumber, err)
		} else {
			processedPages[i].ImagePath = imagePath
		}
	}

	log.Printf("Successfully processed %d pages (simulated)", len(processedPages))
	return processedPages, nil
}

// extractTablesAdvanced uses sophisticated pattern matching to identify tables
func (p *PDFProcessor) extractTablesAdvanced(text string, pageNum int) []Table {
	var tables []Table

	// Split text into lines
	lines := strings.Split(text, "\n")

	// Table detection patterns
	tablePatterns := []*regexp.Regexp{
		regexp.MustCompile(`^\s*[A-Z][a-z]+\s+[0-9]+\s+[A-Za-z]+\s*$`), // "Metric 4.2 seconds"
		regexp.MustCompile(`^\s*[A-Z][a-z]+\s+[A-Za-z]+\s*$`),          // "City driving"
		regexp.MustCompile(`^\s*[0-9]+\s+[A-Za-z%]+\s*$`),              // "28 mpg"
		regexp.MustCompile(`^\s*[A-Z][a-z]+\s*:\s*[0-9.]+`),            // "Power: 395"
	}

	var currentTable [][]string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line matches table patterns
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
				// Check if columns look like table data
				hasNumbers := false
				hasText := false
				for _, col := range columns {
					if regexp.MustCompile(`^[0-9.]+$`).MatchString(col) {
						hasNumbers = true
					} else if regexp.MustCompile(`^[A-Za-z]+$`).MatchString(col) {
						hasText = true
					}
				}
				if hasNumbers && hasText {
					isTableRow = true
				}
			}
		}

		if isTableRow {
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
						X: 0,
						Y: 0,
						W: 595,
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
				X: 0,
				Y: 0,
				W: 595,
				H: float64(len(currentTable) * 12),
			},
		}
		tables = append(tables, table)
	}

	return tables
}

// extractFiguresAdvanced identifies figures, charts, and images in the document
func (p *PDFProcessor) extractFiguresAdvanced(text string, pageNum int) []Figure {
	var figures []Figure

	// Figure detection patterns
	figurePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s*Figure\s+[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Fig\.\s*[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Chart\s+[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Graph\s+[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Image\s+[0-9]+[:\s]`),
		regexp.MustCompile(`(?i)^\s*Photo\s+[0-9]+[:\s]`),
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for figure patterns
		for _, pattern := range figurePatterns {
			if pattern.MatchString(line) {
				// Extract caption
				caption := strings.TrimSpace(line)

				figure := Figure{
					Caption:   caption,
					PageNum:   pageNum,
					ImagePath: "", // Will be filled by image extraction
					BBox: BoundingBox{
						X: 50,  // Approximate figure position
						Y: 100, // Approximate figure position
						W: 495, // Approximate figure width
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
func (p *PDFProcessor) generatePageImage(imagePath string, pageNum int, docID string) (string, error) {
	// Create document-specific directory
	docDir := filepath.Join(p.outputDir, "pages", docID)
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create document directory: %w", err)
	}

	// Generate image filename
	finalImagePath := filepath.Join(docDir, fmt.Sprintf("%d.png", pageNum))

	// Create a simple placeholder image
	if err := p.createPlaceholderImage(finalImagePath, pageNum); err != nil {
		return "", fmt.Errorf("failed to create placeholder image: %w", err)
	}

	return finalImagePath, nil
}

// createPlaceholderImage creates a simple placeholder image for development
func (p *PDFProcessor) createPlaceholderImage(imagePath string, pageNum int) error {
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

// ExtractTextFromContent extracts text from content
func (p *PDFProcessor) ExtractTextFromContent(content string) string {
	return strings.TrimSpace(content)
}

// ExtractTablesFromText extracts table structures from text
func (p *PDFProcessor) ExtractTablesFromText(text string, pageNum int) []Table {
	return p.extractTablesAdvanced(text, pageNum)
}

// ExtractFiguresFromText extracts figure information from text
func (p *PDFProcessor) ExtractFiguresFromText(text string, pageNum int) []Figure {
	return p.extractFiguresAdvanced(text, pageNum)
}

// Close cleans up resources
func (p *PDFProcessor) Close() error {
	return nil
}
