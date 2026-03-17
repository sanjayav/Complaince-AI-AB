 import uuid
 from datetime import datetime
 from fastapi import APIRouter, BackgroundTasks, Depends, HTTPException
 from sqlalchemy.orm import Session

from app.deps import Identity, get_current_identity, get_tenant_scoped_db
from app.models import Document, DocumentStatus, DocumentType
from app.schemas import IngestRequest, PurgeRequest, PurgeResponse
from app.rag.ingestion import ingest_s3_document
from app.audit import write_audit


 router = APIRouter()


 @router.post("/ingest")
 def ingest(
     payload: IngestRequest,
     background: BackgroundTasks,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
 ):
     doc = Document(
         id=uuid.uuid4(),
         tenant_id=identity.tenant_id,
         s3_key=payload.s3_key,
         status=DocumentStatus.indexing,
        doc_type=payload.doc_type or DocumentType.report,
        framework=payload.framework,
        jurisdiction=payload.jurisdiction,
        version=payload.version,
         metadata=payload.document_metadata or {},
         created_at=datetime.utcnow(),
         updated_at=datetime.utcnow(),
     )
     db.add(doc)
    write_audit(
        db=db,
        tenant_id=identity.tenant_id,
        user_id=identity.user_id,
        action="document.ingest_queued",
        entity_type="document",
        entity_id=str(doc.id),
        metadata={"s3_key": payload.s3_key},
    )
    db.commit()

     background.add_task(ingest_s3_document, doc_id=str(doc.id))
     return {"documentId": str(doc.id), "status": "queued"}


@router.delete("/{document_id}")
def delete_document(
    document_id: uuid.UUID,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
    doc = (
        db.query(Document)
        .filter(Document.id == document_id, Document.tenant_id == identity.tenant_id)
        .first()
    )
    if not doc:
        return {"status": "not_found"}

    db.delete(doc)  # cascades to chunks
    write_audit(
        db=db,
        tenant_id=identity.tenant_id,
        user_id=identity.user_id,
        action="document.deleted",
        entity_type="document",
        entity_id=str(document_id),
        metadata={"s3_key": doc.s3_key, "sha256": doc.sha256},
    )
    db.commit()
    return {"status": "deleted"}


@router.post("/purge", response_model=PurgeResponse)
def purge_documents(
    payload: PurgeRequest,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
    q = db.query(Document).filter(Document.tenant_id == identity.tenant_id)
    f = payload.filters or {}
    if f.get("doc_type"):
        q = q.filter(Document.doc_type == f["doc_type"])
    if f.get("framework"):
        q = q.filter(Document.framework == f["framework"])
    if f.get("jurisdiction"):
        q = q.filter(Document.jurisdiction == f["jurisdiction"])
    if f.get("version"):
        q = q.filter(Document.version == f["version"])

    matched = q.count()
    deleted = 0
    if not payload.dryRun and matched > 0:
        # delete in bulk; cascades to chunks
        # using instances for audit per deletion could be heavy; do lightweight audit summary
        q.delete(synchronize_session=False)
        write_audit(
            db=db,
            tenant_id=identity.tenant_id,
            user_id=identity.user_id,
            action="document.purged",
            entity_type="document",
            entity_id=None,
            metadata={"matched": matched, "filters": f},
        )
        db.commit()
        deleted = matched

    return PurgeResponse(matched_count=matched, deleted_count=deleted)

