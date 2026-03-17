 import uuid
 from typing import Any, Dict, List, Tuple

from sqlalchemy import text
 from sqlalchemy.orm import Session

 from app.models import Chunk
from app.rag.embeddings import EmbeddingModel
from app.config import settings
import json
import re

# Optional LangChain LLM integration
try:  # Best-effort imports to avoid hard dependency during local dev
    from langchain_openai import ChatOpenAI  # type: ignore
    try:
        from langchain.prompts import ChatPromptTemplate  # type: ignore
    except Exception:  # Older LC versions
        ChatPromptTemplate = None  # type: ignore
except Exception:
    ChatOpenAI = None  # type: ignore
    ChatPromptTemplate = None  # type: ignore


def search_chunks(
    db: Session,
    tenant_id: uuid.UUID,
    query_embedding: List[float],
    top_k: int = 5,
    filters: Dict[str, Any] | None = None,
) -> List[Tuple[str, Dict[str, Any]]]:
    # Uses pgvector cosine distance (<->). Ensure vector extension is enabled.
    base_sql = """
        SELECT c.id::text, c.text, c.metadata, (c.embedding <-> :query) AS distance
        FROM chunks c
        JOIN documents d ON d.id = c.document_id
        WHERE c.tenant_id = :tenant_id AND c.embedding IS NOT NULL
    """
    where_clauses: List[str] = []
    params: Dict[str, Any] = {"tenant_id": str(tenant_id), "query": query_embedding, "k": top_k}
    f = filters or {}
    if "doc_type" in f and f["doc_type"]:
        where_clauses.append("d.doc_type::text = :doc_type")
        params["doc_type"] = f["doc_type"]
    if "framework" in f and f["framework"]:
        where_clauses.append("d.framework = :framework")
        params["framework"] = f["framework"]
    if "jurisdiction" in f and f["jurisdiction"]:
        where_clauses.append("d.jurisdiction = :jurisdiction")
        params["jurisdiction"] = f["jurisdiction"]
    if "version" in f and f["version"]:
        where_clauses.append("d.version = :version")
        params["version"] = f["version"]

    if where_clauses:
        base_sql += " AND " + " AND ".join(where_clauses)

    base_sql += " ORDER BY c.embedding <-> :query LIMIT :k"
    sql = text(base_sql)
    rows = db.execute(sql, params).all()
    results: List[Tuple[str, Dict[str, Any]]] = []
     for r in rows:
        chunk_id = r[0]
        chunk_text = r[1]
        meta = r[2] or {}
        distance = float(r[3]) if r[3] is not None else None
        results.append((chunk_text, meta | {"chunk_id": chunk_id, "distance": distance}))
     return results


def search_chunks_keyword(
    db: Session,
    tenant_id: uuid.UUID,
    query_text: str,
    top_k: int = 20,
    filters: Dict[str, Any] | None = None,
) -> List[Tuple[str, Dict[str, Any]]]:
    base_sql = """
        SELECT c.id::text, c.text, c.metadata,
               ts_rank(to_tsvector('english', c.text), plainto_tsquery('english', :q)) AS rank
        FROM chunks c
        JOIN documents d ON d.id = c.document_id
        WHERE c.tenant_id = :tenant_id
          AND to_tsvector('english', c.text) @@ plainto_tsquery('english', :q)
    """
    where_clauses: List[str] = []
    params: Dict[str, Any] = {"tenant_id": str(tenant_id), "q": query_text, "k": top_k}
    f = filters or {}
    if "doc_type" in f and f["doc_type"]:
        where_clauses.append("d.doc_type::text = :doc_type")
        params["doc_type"] = f["doc_type"]
    if "framework" in f and f["framework"]:
        where_clauses.append("d.framework = :framework")
        params["framework"] = f["framework"]
    if "jurisdiction" in f and f["jurisdiction"]:
        where_clauses.append("d.jurisdiction = :jurisdiction")
        params["jurisdiction"] = f["jurisdiction"]
    if "version" in f and f["version"]:
        where_clauses.append("d.version = :version")
        params["version"] = f["version"]
    if where_clauses:
        base_sql += " AND " + " AND ".join(where_clauses)
    base_sql += " ORDER BY rank DESC LIMIT :k"
    rows = db.execute(text(base_sql), params).all()
    results: List[Tuple[str, Dict[str, Any]]] = []
    for r in rows:
        chunk_id = r[0]
        chunk_text = r[1]
        meta = r[2] or {}
        rank = float(r[3]) if r[3] is not None else None
        results.append((chunk_text, meta | {"chunk_id": chunk_id, "rank": rank}))
    return results


def build_prompt(question: str, contexts: List[str]) -> str:
    header = (
        "You are an enterprise ESG compliance assistant. Answer only from the provided context. "
        "Always cite sources using their [chunk_id]. If unsure, say you do not know."
    )
    context_text = "\n\n".join(f"Context {i+1}:\n{{context_{i}}}" for i in range(len(contexts)))
    return f"SYSTEM:\n{header}\n\n{context_text}\n\nUSER:\nQuestion: {{question}}\nAnswer:"


def generate_answer(question: str, contexts: List[str]) -> str:
    # Prefer LangChain + ChatOpenAI if available and configured
    if ChatOpenAI is not None:
        try:
            llm = ChatOpenAI(model="gpt-4o-mini", temperature=0)  # respects OPENAI_API_KEY
            # Build a simple prompt; use ChatPromptTemplate if present, else join manually
            if ChatPromptTemplate is not None:
                template = build_prompt(question, contexts)
                # Map context variables
                variables: Dict[str, Any] = {"question": question}
                for i, ctx in enumerate(contexts):
                    variables[f"context_{i}"] = ctx
                prompt = ChatPromptTemplate.from_template(template)
                chain = prompt | llm
                resp = chain.invoke(variables)  # type: ignore
                text = getattr(resp, "content", None) or str(resp)
                return text
            else:
                # Fallback: manual messages
                system = "You are an enterprise ESG compliance assistant. Answer only from the provided context. Always cite sources using their [chunk_id]. If unsure, say you do not know."
                ctx = "\n\n".join([f"Context {i+1}:\n{c}" for i, c in enumerate(contexts)])
                user = f"{ctx}\n\nQuestion: {question}\nAnswer:"
                resp = llm.invoke([{"role": "system", "content": system}, {"role": "user", "content": user}])  # type: ignore
                text = getattr(resp, "content", None) or str(resp)
                return text
        except Exception:
            pass
    # Deterministic fallback without external calls
    return "Based on the retrieved documents, here is a summarized answer with citations where relevant."


 def answer_question_with_retrieval(
     db: Session, tenant_id: uuid.UUID, question: str, filters: Dict[str, Any]
) -> tuple[str, Dict[str, Any]]:
     embedder = EmbeddingModel()
    q_emb = embedder.embed(question)
    # Fetch candidate pool larger than final top_k to allow reranking/hybrid fusion
    final_top_k = max(1, settings.rag_top_k)
    candidate_k = max(final_top_k * 4, final_top_k)
    vec_results = search_chunks(db=db, tenant_id=tenant_id, query_embedding=q_emb, top_k=candidate_k, filters=filters)
    # Optional hybrid keyword retrieval
    if settings.hybrid_enabled:
        kw_results = search_chunks_keyword(db=db, tenant_id=tenant_id, query_text=question, top_k=candidate_k, filters=filters)
        fused = _hybrid_fuse(vec_results, kw_results, alpha=settings.hybrid_alpha)
        raw_results = fused
    else:
        raw_results = vec_results
    # Optional LLM reranking
    results = raw_results
    if settings.rerank_enabled and ChatOpenAI is not None and len(raw_results) > final_top_k:
        try:
            texts = [t for (t, _m) in raw_results]
            rankings = _llm_rerank(question, texts, top_k=final_top_k)
            # rankings is list of indices; reorder accordingly
            results = [raw_results[i] for i in rankings if 0 <= i < len(raw_results)]
        except Exception:
            results = raw_results[:final_top_k]
    else:
        results = raw_results[:final_top_k]
    contexts = [t for (t, _m) in results]
    answer = generate_answer(question, contexts)
     citations = {
         "chunks": [m for (_t, m) in results],
         "top_k": len(results),
     }
     return answer, citations


def _hybrid_fuse(
    vec_results: List[Tuple[str, Dict[str, Any]]],
    kw_results: List[Tuple[str, Dict[str, Any]]],
    alpha: float = 0.7,
) -> List[Tuple[str, Dict[str, Any]]]:
    # Normalize scores and combine by chunk_id
    alpha = max(0.0, min(1.0, alpha))
    # Vector: distance -> score
    vec_scores: Dict[str, float] = {}
    for _t, m in vec_results:
        dist = m.get("distance")
        if dist is None:
            continue
        score = 1.0 / (1.0 + float(dist))
        vec_scores[m["chunk_id"]] = max(vec_scores.get(m["chunk_id"], 0.0), score)
    # Keyword: rank -> normalized score
    kw_scores_raw: Dict[str, float] = {m["chunk_id"]: float(m.get("rank", 0.0)) for _t, m in kw_results}
    max_rank = max(kw_scores_raw.values()) if kw_scores_raw else 1.0
    kw_scores = {k: (v / max_rank) for k, v in kw_scores_raw.items()} if max_rank > 0 else kw_scores_raw

    # Union all candidate ids
    ids = set(list(vec_scores.keys()) + list(kw_scores.keys()))
    combined: List[Tuple[str, Dict[str, Any], float]] = []
    # Build a lookup for text/meta by id from either list
    meta_by_id: Dict[str, Tuple[str, Dict[str, Any]]] = {}
    for t, m in vec_results:
        meta_by_id[m["chunk_id"]] = (t, m)
    for t, m in kw_results:
        if m["chunk_id"] not in meta_by_id:
            meta_by_id[m["chunk_id"]] = (t, m)

    for cid in ids:
        v = vec_scores.get(cid, 0.0)
        k = kw_scores.get(cid, 0.0)
        fused = alpha * v + (1.0 - alpha) * k
        t, m = meta_by_id[cid]
        combined.append((t, m, fused))
    combined.sort(key=lambda x: x[2], reverse=True)
    return [(t, m) for (t, m, _s) in combined]


def _llm_rerank(question: str, candidates: List[str], top_k: int) -> List[int]:
    """Return indices of top_k candidates ranked by relevance using LLM."""
    if ChatOpenAI is None:
        return list(range(min(top_k, len(candidates))))
    llm = ChatOpenAI(model="gpt-4o-mini", temperature=0)
    numbered = "\n\n".join([f"[{i}]\n{c}" for i, c in enumerate(candidates)])
    instr = (
        "You are ranking chunks for relevance to an ESG compliance question. "
        "Return strictly a JSON object with key 'ranking' as a list of objects with 'index' and 'score' (0-100). "
        "Do not include any extra text."
    )
    user = f"Question: {question}\n\nChunks:\n{numbered}\n\nReturn top {top_k} only."
    resp = llm.invoke([{"role": "system", "content": instr}, {"role": "user", "content": user}])  # type: ignore
    text = getattr(resp, "content", None) or str(resp)
    # Extract JSON payload
    m = re.search(r"\{[\s\S]*\}", text)
    payload = text if m is None else m.group(0)
    data = json.loads(payload)
    items = data.get("ranking", [])
    items_sorted = sorted(items, key=lambda x: x.get("score", 0), reverse=True)
    return [int(x.get("index", 0)) for x in items_sorted[:top_k]]

