#!/usr/bin/env python3
import os
import sys
import uuid

from sqlalchemy.orm import Session

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.database import SessionLocal  # noqa: E402
from app.models import Conversation, Message, Document, Chunk, Setting, AuditLog  # noqa: E402


def env_bool(name: str, default: bool) -> bool:
    raw = os.getenv(name)
    if raw is None:
        return default
    return raw.lower() in ("1", "true", "yes", "y")


def main() -> None:
    dry_run = env_bool("DRY_RUN", True)
    tenant_id_raw = os.getenv("TENANT_ID")
    tenant_id = None
    if tenant_id_raw:
        try:
            tenant_id = uuid.UUID(tenant_id_raw)
        except Exception:
            print("Invalid TENANT_ID; must be UUID", file=sys.stderr)
            sys.exit(1)

    db: Session = SessionLocal()
    try:
        def scoped(q):
            return q.filter_by(tenant_id=tenant_id) if tenant_id else q

        counts = {
            "chunks": scoped(db.query(Chunk)).count(),
            "documents": scoped(db.query(Document)).count(),
            "messages": scoped(db.query(Message)).count(),
            "conversations": scoped(db.query(Conversation)).count(),
            "audit_logs": scoped(db.query(AuditLog)).count(),
            "settings": scoped(db.query(Setting)).count(),
        }

        print({"dry_run": dry_run, "tenant_id": str(tenant_id) if tenant_id else None, "counts": counts})

        if dry_run:
            return

        # Delete in safe order (children first where no ON DELETE CASCADE)
        scoped(db.query(Chunk)).delete(synchronize_session=False)
        scoped(db.query(Document)).delete(synchronize_session=False)
        scoped(db.query(Message)).delete(synchronize_session=False)
        scoped(db.query(Conversation)).delete(synchronize_session=False)
        scoped(db.query(AuditLog)).delete(synchronize_session=False)
        scoped(db.query(Setting)).delete(synchronize_session=False)

        db.commit()
        print({"deleted": counts})
    finally:
        db.close()


if __name__ == "__main__":
    main()


