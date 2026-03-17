 from datetime import datetime
 from typing import Any, Optional

 from sqlalchemy.orm import Session

 from app.models import AuditLog


 def write_audit(
     db: Session,
     tenant_id,
     user_id,
     action: str,
     entity_type: str,
     entity_id: Optional[str] = None,
     metadata: Optional[dict[str, Any]] = None,
) -> None:
     log = AuditLog(
         tenant_id=tenant_id,
         user_id=user_id,
         action=action,
         entity_type=entity_type,
         entity_id=entity_id,
         metadata=metadata or {},
         created_at=datetime.utcnow(),
     )
     db.add(log)

