 from datetime import datetime
 import uuid
 from fastapi import APIRouter, Depends, HTTPException, status
 from sqlalchemy.orm import Session

from app.deps import Identity, get_current_identity, get_tenant_scoped_db
 from app.models import OrgMembership, Role, User
 from app.schemas import AddMemberRequest, MemberItem, MembersListResponse
from app.security import hash_password
from app.audit import write_audit


 router = APIRouter()


 def require_role(identity: Identity, roles: list[Role]):
     if identity.role not in roles:
         raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Insufficient role")


 @router.post("/members")
 def add_member(
     payload: AddMemberRequest,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
 ):
     require_role(identity, [Role.owner, Role.admin])

     user = db.query(User).filter(User.email == payload.email).first()
     if user is None:
         user = User(id=uuid.uuid4(), email=payload.email, password_hash=None, name=None, created_at=datetime.utcnow())
         db.add(user)
         db.flush()

     existing = (
         db.query(OrgMembership)
         .filter(OrgMembership.user_id == user.id, OrgMembership.tenant_id == identity.tenant_id)
         .first()
     )
     if existing:
         raise HTTPException(status_code=400, detail="User already a member")

     membership = OrgMembership(
         id=uuid.uuid4(),
         tenant_id=identity.tenant_id,
         user_id=user.id,
         role=payload.role,
         created_at=datetime.utcnow(),
     )
     db.add(membership)
    write_audit(
        db=db,
        tenant_id=identity.tenant_id,
        user_id=identity.user_id,
        action="org.member_added",
        entity_type="user",
        entity_id=str(user.id),
        metadata={"role": payload.role.value},
    )
    db.commit()
     return {"status": "ok"}


 @router.get("/members", response_model=MembersListResponse)
def list_members(identity: Identity = Depends(get_current_identity), db: Session = Depends(get_tenant_scoped_db)):
     rows = (
         db.query(OrgMembership, User)
         .join(User, User.id == OrgMembership.user_id)
         .filter(OrgMembership.tenant_id == identity.tenant_id)
         .all()
     )
     items = [
         MemberItem(
             user_id=u.id,
             email=u.email,
             name=u.name,
             role=m.role,
             created_at=m.created_at,
         )
         for (m, u) in rows
     ]
     return MembersListResponse(items=items)

