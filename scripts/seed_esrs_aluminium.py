#!/usr/bin/env python3
import os
import sys
import uuid
from datetime import datetime
from typing import List, Dict

from sqlalchemy.orm import Session

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.database import SessionLocal  # noqa: E402
from app.models import Document, DocumentStatus, DocumentType, Chunk, Tenant  # noqa: E402
from app.rag.embeddings import EmbeddingModel  # noqa: E402


CLAUSES: List[Dict[str, str]] = [
    {
        "clause_id": "ESRS E1-5 para 23",
        "heading": "Scope 3 category 1 – Purchased goods (Aluminium)",
        "text": (
            "Undertakings using primary or recycled aluminium in products SHALL disclose Scope 3 category 1 "
            "(Purchased goods and services) GHG emissions. Disclosures SHALL include methodology for supplier-"
            "specific emission factors, recycled content share, and assurance over data quality."
        ),
    },
    {
        "clause_id": "ESRS E2-6 para 12",
        "heading": "Resource use and circularity (Aluminium content)",
        "text": (
            "Undertakings SHALL disclose mass of aluminium input by source (primary vs recycled), scrap recovery rate, "
            "and measures to increase closed-loop recycling within their value chain."
        ),
    },
    {
        "clause_id": "ESRS E1-6 para 30",
        "heading": "Transition plan targets for hard-to-abate materials",
        "text": (
            "Transition plans SHALL include measurable targets to reduce embodied emissions of aluminium, including supplier engagement, "
            "low-carbon smelting procurement (renewable electricity), and design-for-recycling commitments."
        ),
    },
    {
        "clause_id": "ESRS E1 AR 16",
        "heading": "Supplier data and verification",
        "text": (
            "When supplier-specific factors are used for aluminium, undertakings SHOULD disclose verification approach, frequency, and coverage, "
            "including third-party assurance or chain-of-custody certifications where applicable."
        ),
    },
    {
        "clause_id": "ESRS E5-3 para 18",
        "heading": "Design and material efficiency",
        "text": (
            "Undertakings SHALL disclose product design measures that reduce aluminium mass, improve repairability, and facilitate material separation at end-of-life."
        ),
    },
]


def get_env(name: str, default: str | None = None) -> str:
    val = os.getenv(name, default)
    if val is None:
        raise SystemExit(f"Missing required env: {name}")
    return val


def main() -> None:
    tenant_id_str = os.getenv("TENANT_ID")
    tenant_name = os.getenv("TENANT_NAME")
    framework = os.getenv("FRAMEWORK", "ESRS")
    jurisdiction = os.getenv("JURISDICTION", "EU")
    version = os.getenv("VERSION", "v1.2")

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

        # Idempotency: if a seed-tagged document exists for this tenant+version, skip
        existing = (
            db.query(Document)
            .filter(
                Document.tenant_id == tenant_id,
                Document.doc_type == DocumentType.regulation,
                Document.framework == framework,
                Document.jurisdiction == jurisdiction,
                Document.version == version,
            )
            .first()
        )
        if existing:
            print({"status": "exists", "document_id": str(existing.id)})
            return

        doc = Document(
            id=uuid.uuid4(),
            tenant_id=tenant_id,
            s3_key=f"seed/esrs/aluminium/{version}.txt",
            status=DocumentStatus.indexed,
            doc_type=DocumentType.regulation,
            framework=framework,
            jurisdiction=jurisdiction,
            version=version,
            metadata={"seed": "esrs_aluminium"},
            created_at=datetime.utcnow(),
            updated_at=datetime.utcnow(),
        )
        db.add(doc)
        db.flush()

        embedder = EmbeddingModel()
        texts = [f"{c['clause_id']} - {c['heading']}\n{c['text']}" for c in CLAUSES]
        embs = embedder.embed_many(texts)
        dim = len(embs[0]) if embs else 1536

        for c, emb, text in zip(CLAUSES, embs, texts):
            meta = {
                "framework": framework,
                "jurisdiction": jurisdiction,
                "version": version,
                "doc_type": DocumentType.regulation.value,
                "clause_id": c["clause_id"],
                "heading": c["heading"],
            }
            ch = Chunk(
                tenant_id=tenant_id,
                document_id=doc.id,
                text=text,
                page_no=None,
                heading=c["heading"],
                metadata=meta,
                dimensions=dim,
                embedding=emb,
            )
            db.add(ch)

        db.commit()
        print({"status": "created", "document_id": str(doc.id), "chunks": len(CLAUSES)})
    finally:
        db.close()


if __name__ == "__main__":
    main()


