-- Add doc_type enum and regulatory fields to documents
DO $$ BEGIN
  CREATE TYPE doc_type_enum AS ENUM ('regulation','company_policy','report','supplier_data');
EXCEPTION WHEN duplicate_object THEN null; END $$;

ALTER TABLE documents
  ADD COLUMN IF NOT EXISTS doc_type doc_type_enum NOT NULL DEFAULT 'report',
  ADD COLUMN IF NOT EXISTS framework VARCHAR(128),
  ADD COLUMN IF NOT EXISTS jurisdiction VARCHAR(128),
  ADD COLUMN IF NOT EXISTS version VARCHAR(32);

CREATE INDEX IF NOT EXISTS idx_documents_framework ON documents(framework);
CREATE INDEX IF NOT EXISTS idx_documents_jurisdiction ON documents(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_documents_version ON documents(version);

-- Optional: ensure sha256 dedup per tenant
CREATE UNIQUE INDEX IF NOT EXISTS uq_documents_tenant_sha256 ON documents(tenant_id, sha256) WHERE sha256 IS NOT NULL;


