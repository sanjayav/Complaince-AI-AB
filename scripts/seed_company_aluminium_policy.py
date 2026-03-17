#!/usr/bin/env python3
import os
import sys
import uuid
from datetime import datetime

from sqlalchemy.orm import Session

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.database import SessionLocal  # noqa: E402
from app.models import Document, DocumentStatus, DocumentType, Chunk, Tenant  # noqa: E402
from app.rag.embeddings import EmbeddingModel  # noqa: E402
from app.rag.chunking import simple_chunk  # noqa: E402


def main() -> None:
    tenant_id_str = os.getenv("TENANT_ID")
    tenant_name = os.getenv("TENANT_NAME")
    framework = os.getenv("FRAMEWORK", "ESRS")
    jurisdiction = os.getenv("JURISDICTION", "EU")
    version = os.getenv("VERSION", "v1.2")
    path = os.getenv("PATH", "test-data/company_aluminium_policy.txt")

    if not os.path.exists(path):
        raise SystemExit(f"File not found: {path}")

    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    db: Session = SessionLocal()
    try:
        tenant_id = None
        if tenant_id_str:
            tenant_id = uuid.UUID(tenant_id_str)
            t = db.query(Tenant).filter(Tenant.id == tenant_id).first()
            if not t:
                raise SystemExit("TENANT_ID not found")
        elif tenant_name:
            t = db.query(Tenant).filter(Tenant.name == tenant_name).first()
            if not t:
                raise SystemExit("TENANT_NAME not found")
            tenant_id = t.id
        else:
            raise SystemExit("Provide TENANT_ID or TENANT_NAME in env")

        # Idempotency: skip if a matching company_policy doc exists
        existing = (
            db.query(Document)
            .filter(
                Document.tenant_id == tenant_id,
                Document.doc_type == DocumentType.company_policy,
                Document.framework == framework,
                Document.jurisdiction == jurisdiction,
                Document.version == version,
                Document.s3_key == path,
            )
            .first()
        )
        if existing:
            print({"status": "exists", "document_id": str(existing.id)})
            return

        doc = Document(
            id=uuid.uuid4(),
            tenant_id=tenant_id,
            s3_key=path,
            status=DocumentStatus.indexed,
            doc_type=DocumentType.company_policy,
            framework=framework,
            jurisdiction=jurisdiction,
            version=version,
            metadata={"seed": "company_aluminium_policy"},
            created_at=datetime.utcnow(),
            updated_at=datetime.utcnow(),
        )
        db.add(doc)
        db.flush()

        chunks = simple_chunk(content, max_chars=1200, overlap=120)
        embedder = EmbeddingModel()
        embs = embedder.embed_many(chunks)
        dim = len(embs[0]) if embs else 1536

        for text, emb in zip(chunks, embs):
            meta = {
                "framework": framework,
                "jurisdiction": jurisdiction,
                "version": version,
                "doc_type": DocumentType.company_policy.value,
            }
            ch = Chunk(
                tenant_id=tenant_id,
                document_id=doc.id,
                text=text,
                page_no=None,
                heading=None,
                metadata=meta,
                dimensions=dim,
                embedding=emb,
            )
            db.add(ch)

        db.commit()
        print({"status": "created", "document_id": str(doc.id), "chunks": len(chunks)})
    finally:
        db.close()


if __name__ == "__main__":
    main()


