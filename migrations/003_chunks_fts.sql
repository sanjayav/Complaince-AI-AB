-- Full-text search index on chunks.text for hybrid retrieval
CREATE INDEX IF NOT EXISTS idx_chunks_text_fts
ON chunks USING GIN (to_tsvector('english', text));


