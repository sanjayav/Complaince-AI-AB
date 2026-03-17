#!/usr/bin/env python3
import os
import sys
import uuid
from datetime import datetime

from sqlalchemy.orm import Session

sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.database import SessionLocal  # noqa: E402
from app.models import Tenant, User, OrgMembership, Role  # noqa: E402
from app.security import hash_password  # noqa: E402


def get_env(name: str, default: str | None = None) -> str:
    val = os.getenv(name, default)
    if val is None:
        raise SystemExit(f"Missing required env: {name}")
    return val


def provision(tenant_name: str, admin_email: str, admin_password: str) -> None:
    db: Session = SessionLocal()
    try:
        tenant = db.query(Tenant).filter(Tenant.name == tenant_name).first()
        if tenant is None:
            tenant = Tenant(id=uuid.uuid4(), name=tenant_name, created_at=datetime.utcnow())
            db.add(tenant)
            db.flush()

        user = db.query(User).filter(User.email == admin_email).first()
        if user is None:
            user = User(
                id=uuid.uuid4(),
                email=admin_email,
                password_hash=hash_password(admin_password),
                name="Administrator",
                created_at=datetime.utcnow(),
            )
            db.add(user)
            db.flush()
        else:
            # ensure has password
            if not user.password_hash:
                user.password_hash = hash_password(admin_password)
                db.add(user)

        membership = (
            db.query(OrgMembership)
            .filter(OrgMembership.user_id == user.id, OrgMembership.tenant_id == tenant.id)
            .first()
        )
        if membership is None:
            membership = OrgMembership(
                id=uuid.uuid4(),
                tenant_id=tenant.id,
                user_id=user.id,
                role=Role.owner,
                created_at=datetime.utcnow(),
            )
            db.add(membership)

        db.commit()
        print(
            {
                "tenant_id": str(tenant.id),
                "tenant_name": tenant.name,
                "admin_user_id": str(user.id),
                "admin_email": user.email,
                "role": membership.role.value,
            }
        )
    finally:
        db.close()


if __name__ == "__main__":
    tenant_name = get_env("TENANT_NAME", "DEV")
    admin_email = get_env("ADMIN_EMAIL", "admin@example.com")
    admin_password = get_env("ADMIN_PASSWORD", "admin123!")
    provision(tenant_name, admin_email, admin_password)


