package jobs

// Stub for indexing worker

type IndexJob struct {
	DocumentID string
}

func Enqueue(job IndexJob) error {
	// TODO: push to queue or run immediately
	return nil
}
