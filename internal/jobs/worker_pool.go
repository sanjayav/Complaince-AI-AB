package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"jlrdi/internal/rag"
)

// Job represents a document processing job
type Job struct {
	ID       string
	FilePath string
	S3Key    string
	DocType  string
	Priority int
	Created  time.Time
}

// JobResult represents the result of a processing job
type JobResult struct {
	JobID      string
	Success    bool
	Error      error
	Pages      int
	Tables     int
	Figures    int
	Processing time.Duration
}

// WorkerPool manages concurrent document processing
type WorkerPool struct {
	workers    int
	jobQueue   chan Job
	resultChan chan JobResult
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	processor  rag.PDFProcessor
}

// NewWorkerPool creates a new worker pool for document processing
func NewWorkerPool(workers int, processor rag.PDFProcessor) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:    workers,
		jobQueue:   make(chan Job, workers*2),
		resultChan: make(chan JobResult, workers*2),
		ctx:        ctx,
		cancel:     cancel,
		processor:  processor,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers", wp.workers)

	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// Start result processor
	go wp.processResults()
}

// Stop stops the worker pool gracefully
func (wp *WorkerPool) Stop() {
	log.Println("Stopping worker pool...")
	wp.cancel()
	close(wp.jobQueue)
	close(wp.resultChan)
	wp.wg.Wait()
	log.Println("Worker pool stopped")
}

// SubmitJob submits a job to the worker pool
func (wp *WorkerPool) SubmitJob(job Job) error {
	select {
	case wp.jobQueue <- job:
		log.Printf("Job %s submitted to worker pool", job.ID)
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is stopped")
	default:
		return fmt.Errorf("worker pool is full")
	}
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case job, ok := <-wp.jobQueue:
			if !ok {
				log.Printf("Worker %d shutting down", id)
				return
			}

			log.Printf("Worker %d processing job %s", id, job.ID)
			start := time.Now()

			// Process the job
			result := wp.processJob(job)
			result.Processing = time.Since(start)

			// Send result
			select {
			case wp.resultChan <- result:
				log.Printf("Worker %d completed job %s in %v", id, job.ID, result.Processing)
			case <-wp.ctx.Done():
				return
			}

		case <-wp.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		}
	}
}

// processJob processes a single job
func (wp *WorkerPool) processJob(job Job) JobResult {
	result := JobResult{
		JobID:   job.ID,
		Success: false,
	}

	// Process the PDF
	processedPages, err := wp.processor.ProcessPDF(wp.ctx, job.FilePath, job.ID)
	if err != nil {
		result.Error = fmt.Errorf("failed to process PDF: %w", err)
		return result
	}

	// Count extracted content
	var totalTables, totalFigures int
	for _, page := range processedPages {
		totalTables += len(page.Tables)
		totalFigures += len(page.Figures)
	}

	result.Success = true
	result.Pages = len(processedPages)
	result.Tables = totalTables
	result.Figures = totalFigures

	return result
}

// processResults processes job results
func (wp *WorkerPool) processResults() {
	for {
		select {
		case result, ok := <-wp.resultChan:
			if !ok {
				return
			}

			if result.Success {
				log.Printf("Job %s completed successfully: %d pages, %d tables, %d figures",
					result.JobID, result.Pages, result.Tables, result.Figures)
			} else {
				log.Printf("Job %s failed: %v", result.JobID, result.Error)
			}

		case <-wp.ctx.Done():
			return
		}
	}
}

// GetStats returns current worker pool statistics
func (wp *WorkerPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"workers":    wp.workers,
		"queue_size": len(wp.jobQueue),
		"active":     wp.workers,
		"status":     "running",
	}
}
