 ## Compliance Assistant Backend (Multi-tenant SaaS)

 FastAPI backend with multi-tenant RBAC, JWT auth, Postgres + pgvector, S3-powered ingestion with LangChain, and chat/RAG retrieval.

 ### Quickstart
 1) Install dependencies
 
 ```bash
 python3 -m venv .venv && source .venv/bin/activate
 pip install -r requirements.txt
 ```
 
 2) Provision Postgres with pgvector
 - Create a database and enable the `vector` extension:
 
 ```sql
 CREATE EXTENSION IF NOT EXISTS vector;
 ```
 
 3) Configure environment
 - Copy `.env.example` to `.env` and fill values.
 
 4) Run migrations (SQL)
 
 ```bash
 psql "$DATABASE_URL" -f migrations/001_init.sql
 ```
 
 5) Run the API
 
 ```bash
 uvicorn app.main:app --reload
 ```
 
 ### Key Endpoints (initial)
 - POST `/auth/login` – email/password → JWT, tenant, user
 - GET `/conversations` – paginated list
 - GET `/conversations/{id}/messages` – message history
 - POST `/chat` – ask a question, persists messages
 - POST `/documents/ingest` – index an S3 object into vector DB
 - POST `/org/members` – add user to tenant
 - GET `/org/members` – list members
 
 ### Notes
 - RLS policies are included in `migrations/001_init.sql` (require setting `app.current_tenant` per connection for enforcement). The app also filters by `tenant_id` at the application layer.
 - Embeddings default to OpenAI if `OPENAI_API_KEY` is configured; otherwise a local embedding model can be wired later.

