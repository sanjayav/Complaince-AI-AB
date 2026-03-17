package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations executes database migrations
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		createDocumentsTable,
		alterDocumentsAddTenantID,
		alterDocumentsAddS3URL,
		createPagesTable,
		createCellsTable,
		createFiguresTable,
		createTextChunksTable,
		createAnswersTable,
		createQATasksTable,
		createIndexes,
	}

	for i, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}

const createDocumentsTable = `
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename VARCHAR(255) NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'pending',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

const alterDocumentsAddTenantID = `
ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(100) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_documents_tenant_id ON documents(tenant_id);
`

const alterDocumentsAddS3URL = `
ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS s3_url TEXT;
CREATE INDEX IF NOT EXISTS idx_documents_s3_url ON documents(s3_url);
`

const createPagesTable = `
CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_number INTEGER NOT NULL,
    s3_image_key VARCHAR(500),
    width INTEGER,
    height INTEGER,
    text_content TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(document_id, page_number)
);
`

const createCellsTable = `
CREATE TABLE IF NOT EXISTS cells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    cell_type VARCHAR(50) NOT NULL, -- 'header', 'data', 'total'
    row_index INTEGER,
    col_index INTEGER,
    bbox_x FLOAT NOT NULL,
    bbox_y FLOAT NOT NULL,
    bbox_width FLOAT NOT NULL,
    bbox_height FLOAT NOT NULL,
    text_content TEXT NOT NULL,
    confidence FLOAT,
    embedding_id VARCHAR(100), -- Qdrant vector ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

const createFiguresTable = `
CREATE TABLE IF NOT EXISTS figures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    figure_type VARCHAR(50), -- 'chart', 'graph', 'diagram', 'image'
    caption TEXT,
    bbox_x FLOAT NOT NULL,
    bbox_y FLOAT NOT NULL,
    bbox_width FLOAT NOT NULL,
    bbox_height FLOAT NOT NULL,
    s3_image_key VARCHAR(500),
    embedding_id VARCHAR(100), -- Qdrant vector ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

const createTextChunksTable = `
CREATE TABLE IF NOT EXISTS text_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    chunk_type VARCHAR(50) NOT NULL, -- 'paragraph', 'section', 'list'
    text_content TEXT NOT NULL,
    bbox_x FLOAT,
    bbox_y FLOAT,
    bbox_width FLOAT,
    bbox_height FLOAT,
    embedding_id VARCHAR(100), -- Qdrant vector ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

const createAnswersTable = `
CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    citations JSONB NOT NULL, -- Array of citation objects
    model_used VARCHAR(100),
    tokens_used INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

const createQATasksTable = `
CREATE TABLE IF NOT EXISTS qa_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_type VARCHAR(50) NOT NULL, -- 'classification', 'figure_linking', 'citation_review'
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    entity_id VARCHAR(100) NOT NULL, -- UUID of the entity being reviewed
    entity_type VARCHAR(50) NOT NULL, -- 'cell', 'figure', 'chunk'
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'approved', 'rejected'
    assigned_to VARCHAR(100),
    reviewed_by VARCHAR(100),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    feedback TEXT,
    corrections JSONB, -- Store corrections for feedback loop
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_uploaded_at ON documents(uploaded_at);
CREATE INDEX IF NOT EXISTS idx_pages_document_id ON pages(document_id);
CREATE INDEX IF NOT EXISTS idx_cells_document_id ON cells(document_id);
CREATE INDEX IF NOT EXISTS idx_cells_page_id ON cells(page_id);
CREATE INDEX IF NOT EXISTS idx_figures_document_id ON figures(document_id);
CREATE INDEX IF NOT EXISTS idx_figures_page_id ON figures(page_id);
CREATE INDEX IF NOT EXISTS idx_text_chunks_document_id ON text_chunks(document_id);
CREATE INDEX IF NOT EXISTS idx_text_chunks_page_id ON text_chunks(page_id);
CREATE INDEX IF NOT EXISTS idx_qa_tasks_status ON qa_tasks(status);
CREATE INDEX IF NOT EXISTS idx_qa_tasks_assigned_to ON qa_tasks(assigned_to);
`
