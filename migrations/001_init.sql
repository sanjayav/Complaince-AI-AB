 -- Enable extensions
 CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
 CREATE EXTENSION IF NOT EXISTS vector;

 -- Enums
 DO $$ BEGIN
   CREATE TYPE role_enum AS ENUM ('owner','admin','analyst','viewer');
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE TYPE message_role_enum AS ENUM ('user','assistant');
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE TYPE doc_status_enum AS ENUM ('uploaded','indexing','indexed','failed');
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 -- Tables
 CREATE TABLE IF NOT EXISTS tenants (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   name VARCHAR(255) NOT NULL UNIQUE,
   created_at TIMESTAMP NOT NULL DEFAULT now()
 );

 CREATE TABLE IF NOT EXISTS users (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   email VARCHAR(255) NOT NULL UNIQUE,
   password_hash VARCHAR(255),
   name VARCHAR(255),
   created_at TIMESTAMP NOT NULL DEFAULT now()
 );

 CREATE TABLE IF NOT EXISTS org_memberships (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
   user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
   role role_enum NOT NULL DEFAULT 'viewer',
   created_at TIMESTAMP NOT NULL DEFAULT now(),
   CONSTRAINT uq_membership_tenant_user UNIQUE (tenant_id, user_id)
 );

 CREATE TABLE IF NOT EXISTS conversations (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
   user_id UUID REFERENCES users(id) ON DELETE SET NULL,
   name VARCHAR(255) NOT NULL,
   created_at TIMESTAMP NOT NULL DEFAULT now(),
   deleted_at TIMESTAMP
 );

 CREATE TABLE IF NOT EXISTS messages (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
   conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
   user_id UUID REFERENCES users(id) ON DELETE SET NULL,
   role message_role_enum NOT NULL,
   content TEXT NOT NULL,
   citations JSONB,
   created_at TIMESTAMP NOT NULL DEFAULT now()
 );

 CREATE TABLE IF NOT EXISTS documents (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
   s3_key VARCHAR(1024) NOT NULL,
   status doc_status_enum NOT NULL DEFAULT 'uploaded',
   sha256 VARCHAR(64),
   metadata JSONB,
   created_at TIMESTAMP NOT NULL DEFAULT now(),
   updated_at TIMESTAMP NOT NULL DEFAULT now()
 );

 CREATE TABLE IF NOT EXISTS chunks (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
   document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
   text TEXT NOT NULL,
   page_no INT,
   heading VARCHAR(512),
   metadata JSONB,
   dimensions INT NOT NULL DEFAULT 1536 CHECK (dimensions > 0),
   embedding VECTOR(1536),
   created_at TIMESTAMP NOT NULL DEFAULT now()
 );

 CREATE TABLE IF NOT EXISTS settings (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
   model_name VARCHAR(255) NOT NULL DEFAULT 'gpt-4o-mini',
   retrieval_top_k INT NOT NULL DEFAULT 5,
   safety_level VARCHAR(32) NOT NULL DEFAULT 'standard',
   created_at TIMESTAMP NOT NULL DEFAULT now(),
   updated_at TIMESTAMP NOT NULL DEFAULT now()
 );

 CREATE TABLE IF NOT EXISTS audit_logs (
   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
   tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
   user_id UUID REFERENCES users(id) ON DELETE SET NULL,
   action VARCHAR(64) NOT NULL,
   entity_type VARCHAR(64) NOT NULL,
   entity_id VARCHAR(64),
   metadata JSONB,
   created_at TIMESTAMP NOT NULL DEFAULT now()
 );

 -- Indexes
 CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id);
 CREATE INDEX IF NOT EXISTS idx_messages_tenant ON messages(tenant_id);
 CREATE INDEX IF NOT EXISTS idx_chunks_tenant ON chunks(tenant_id);
 CREATE INDEX IF NOT EXISTS idx_docs_tenant ON documents(tenant_id);

 -- RLS setup (optional, enable in prod)
 ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
 ALTER TABLE users ENABLE ROW LEVEL SECURITY;
 ALTER TABLE org_memberships ENABLE ROW LEVEL SECURITY;
 ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;
 ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
 ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
 ALTER TABLE chunks ENABLE ROW LEVEL SECURITY;
 ALTER TABLE settings ENABLE ROW LEVEL SECURITY;
 ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

 -- Policy uses application-defined setting: app.current_tenant
 -- Ensure each connection sets: SELECT set_config('app.current_tenant', '<tenant_uuid>', true);

 DO $$ BEGIN
   CREATE POLICY tenant_isolation_tenants ON tenants
     USING (id::text = current_setting('app.current_tenant', true));
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE POLICY tenant_isolation_org_memberships ON org_memberships
     USING (tenant_id::text = current_setting('app.current_tenant', true));
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE POLICY tenant_isolation_conversations ON conversations
     USING (tenant_id::text = current_setting('app.current_tenant', true));
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE POLICY tenant_isolation_messages ON messages
     USING (tenant_id::text = current_setting('app.current_tenant', true));
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE POLICY tenant_isolation_documents ON documents
     USING (tenant_id::text = current_setting('app.current_tenant', true));
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE POLICY tenant_isolation_chunks ON chunks
     USING (tenant_id::text = current_setting('app.current_tenant', true));
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE POLICY tenant_isolation_settings ON settings
     USING (tenant_id::text = current_setting('app.current_tenant', true));
 EXCEPTION WHEN duplicate_object THEN null; END $$;

 DO $$ BEGIN
   CREATE POLICY tenant_isolation_audit_logs ON audit_logs
     USING (tenant_id::text = current_setting('app.current_tenant', true));
 EXCEPTION WHEN duplicate_object THEN null; END $$;

