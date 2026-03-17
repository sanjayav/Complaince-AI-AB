 import uuid
 from dataclasses import dataclass
 from typing import Optional

from fastapi import Depends, HTTPException, status
from sqlalchemy import text
 from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
 from sqlalchemy.orm import Session

 from app.database import get_db
 from app.models import OrgMembership, Role, User
 from app.security import decode_access_token


 http_bearer = HTTPBearer(auto_error=False)


 @dataclass
 class Identity:
     user_id: uuid.UUID
     tenant_id: uuid.UUID
     role: Role


 def get_current_identity(
     creds: Optional[HTTPAuthorizationCredentials] = Depends(http_bearer),
     db: Session = Depends(get_db),
 ) -> Identity:
     if creds is None or not creds.scheme.lower() == "bearer":
         raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Not authenticated")

     try:
         payload = decode_access_token(creds.credentials)
     except Exception:
         raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")

     user_id = uuid.UUID(payload.get("userId"))
     tenant_id = uuid.UUID(payload.get("tenantId"))
     role_str = payload.get("role")
     try:
         role = Role(role_str)
     except Exception:
         raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Invalid role")

     # Optional: verify membership still exists
     membership = (
         db.query(OrgMembership)
         .filter(OrgMembership.user_id == user_id, OrgMembership.tenant_id == tenant_id)
         .first()
     )
     if membership is None:
         raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Membership not found")

     return Identity(user_id=user_id, tenant_id=tenant_id, role=role)


def get_tenant_scoped_db(
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_db),
) -> Session:
    # Set per-connection tenant for RLS policies
    try:
        db.execute(text("select set_config('app.current_tenant', :tenant, true)"), {"tenant": str(identity.tenant_id)})
    except Exception:
        # If set_config fails (e.g., not Postgres), continue without DB-level RLS
        pass
    return db

