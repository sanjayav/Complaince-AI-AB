import uuid
from typing import Dict, List

from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from app.deps import Identity, get_current_identity, get_tenant_scoped_db
from app.models import Chunk, DocumentType
from app.schemas import (
    MappingRequest,
    MappingResponse,
    MappingItem,
    MatchItem,
    ScoreRequest,
    ScoreResponse,
)
from app.rag.embeddings import EmbeddingModel
from app.rag.retrieval import search_chunks
from app.audit import write_audit


router = APIRouter()


@router.post("/map", response_model=MappingResponse)
def map_document_to_regulations(
    payload: MappingRequest,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
    # Fetch source (company) chunks
    rows: List[Chunk] = (
        db.query(Chunk)
        .filter(Chunk.tenant_id == identity.tenant_id, Chunk.document_id == payload.documentId)
        .order_by(Chunk.created_at.asc(), Chunk.id.asc())
        .limit(payload.maxSourceChunks)
        .all()
    )

    embedder = EmbeddingModel()
    filters: Dict[str, str] = {"doc_type": DocumentType.regulation.value}
    if payload.filters:
        filters.update({k: v for k, v in payload.filters.items() if v})

    items: List[MappingItem] = []
    for ch in rows:
        q_emb = embedder.embed(ch.text)
        matches = search_chunks(
            db=db,
            tenant_id=identity.tenant_id,
            query_embedding=q_emb,
            top_k=payload.topK,
            filters=filters,
        )
        match_items = [MatchItem(chunk_id=m[1]["chunk_id"], text=t, metadata=m[1]) for (t, m) in matches]
        items.append(
            MappingItem(
                source_chunk_id=str(ch.id),
                source_text=ch.text,
                matches=match_items,
            )
        )

    write_audit(
        db=db,
        tenant_id=identity.tenant_id,
        user_id=identity.user_id,
        action="compliance.mapping_generated",
        entity_type="document",
        entity_id=str(payload.documentId),
        metadata={"items": len(items)},
    )
    db.commit()

    return MappingResponse(items=items)


@router.post("/score", response_model=ScoreResponse)
def score_document_alignment(
    payload: ScoreRequest,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
    # Use a lightweight mapping to compute coverage
    rows: List[Chunk] = (
        db.query(Chunk)
        .filter(Chunk.tenant_id == identity.tenant_id, Chunk.document_id == payload.documentId)
        .order_by(Chunk.created_at.asc(), Chunk.id.asc())
        .limit(200)
        .all()
    )
    total = len(rows)
    if total == 0:
        return ScoreResponse(score=0.0, total_source_chunks=0, covered_source_chunks=0, details=None)

    embedder = EmbeddingModel()
    filters: Dict[str, str] = {"doc_type": DocumentType.regulation.value}
    if payload.framework:
        filters["framework"] = payload.framework
    if payload.jurisdiction:
        filters["jurisdiction"] = payload.jurisdiction
    if payload.version:
        filters["version"] = payload.version

    covered = 0
    for ch in rows:
        q_emb = embedder.embed(ch.text)
        matches = search_chunks(
            db=db,
            tenant_id=identity.tenant_id,
            query_embedding=q_emb,
            top_k=1,
            filters=filters,
        )
        if len(matches) > 0:
            covered += 1

    score = covered / total

    write_audit(
        db=db,
        tenant_id=identity.tenant_id,
        user_id=identity.user_id,
        action="compliance.scored",
        entity_type="document",
        entity_id=str(payload.documentId),
        metadata={"score": score, "total": total, "covered": covered},
    )
    db.commit()

    return ScoreResponse(
        score=score,
        total_source_chunks=total,
        covered_source_chunks=covered,
        details=None,
    )


