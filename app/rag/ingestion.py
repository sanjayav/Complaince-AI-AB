 import hashlib
 import io
 from datetime import datetime
 from typing import List

 import boto3
 import pdfplumber
 from sqlalchemy.orm import Session

 from app.config import settings
 from app.database import SessionLocal
 from app.models import Chunk, Document, DocumentStatus
 from app.rag.chunking import simple_chunk
 from app.rag.embeddings import EmbeddingModel


 def _extract_text_from_pdf_bytes(data: bytes) -> str:
     with pdfplumber.open(io.BytesIO(data)) as pdf:
         texts: List[str] = []
         for page in pdf.pages:
             texts.append(page.extract_text() or "")
     return "\n".join(texts)


 def ingest_s3_document(doc_id: str) -> None:
     db: Session = SessionLocal()
     try:
         document: Document | None = db.query(Document).filter(Document.id == doc_id).first()
         if document is None:
             return

         s3 = boto3.client(
             "s3",
             region_name=settings.aws_region,
             aws_access_key_id=settings.aws_access_key_id,
             aws_secret_access_key=settings.aws_secret_access_key,
         )
         obj = s3.get_object(Bucket=settings.s3_bucket, Key=document.s3_key)
         data = obj["Body"].read()

         sha256 = hashlib.sha256(data).hexdigest()
         document.sha256 = sha256
         document.status = DocumentStatus.indexing
         db.add(document)
         db.flush()

         text = _extract_text_from_pdf_bytes(data)
         chunks = simple_chunk(text)

         embedder = EmbeddingModel()
         embeddings = embedder.embed_many(chunks)

         dim = len(embeddings[0]) if embeddings else 1536

        for chunk_text, emb in zip(chunks, embeddings):
            chunk_meta = (document.metadata or {}).copy()
            # propagate high-level document attributes for filtering
            if getattr(document, "framework", None):
                chunk_meta["framework"] = document.framework
            if getattr(document, "jurisdiction", None):
                chunk_meta["jurisdiction"] = document.jurisdiction
            if getattr(document, "version", None):
                chunk_meta["version"] = document.version
            if getattr(document, "doc_type", None):
                try:
                    chunk_meta["doc_type"] = document.doc_type.value  # enum -> str
                except Exception:
                    pass
             ch = Chunk(
                 tenant_id=document.tenant_id,
                 document_id=document.id,
                 text=chunk_text,
                 page_no=None,
                 heading=None,
                metadata=chunk_meta,
                 dimensions=dim,
                 embedding=emb,
             )
             db.add(ch)

         document.status = DocumentStatus.indexed
         document.updated_at = datetime.utcnow()
         db.commit()
     except Exception:
         try:
             document = db.query(Document).filter(Document.id == doc_id).first()
             if document:
                 document.status = DocumentStatus.failed
                 document.updated_at = datetime.utcnow()
                 db.commit()
         except Exception:
             pass
     finally:
         db.close()

