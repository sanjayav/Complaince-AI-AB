import uuid
from fastapi import APIRouter, Depends, HTTPException, status
import httpx
from jose import jwk, jwt
from jose.utils import base64url_decode
from typing import Any, Dict
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import OrgMembership, User
from app.schemas import LoginRequest, LoginResponse, UserInfo, OidcVerifyRequest
from app.config import settings
from app.security import create_access_token, verify_password

router = APIRouter()

@router.post("/login", response_model=LoginResponse)
def login(payload: LoginRequest, db: Session = Depends(get_db)):
    user = db.query(User).filter(User.email == payload.email).first()
    if user is None or user.password_hash is None:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid credentials")
    if not verify_password(payload.password, user.password_hash):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid credentials")

    membership = db.query(OrgMembership).filter(OrgMembership.user_id == user.id).first()
    if membership is None:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="No tenant membership")

    token = create_access_token(
        {
            "userId": str(user.id),
            "tenantId": str(membership.tenant_id),
            "role": membership.role.value,
        }
    )

    return LoginResponse(
        token=token,
        tenantId=membership.tenant_id,
        role=membership.role,
        user=UserInfo(id=user.id, email=user.email, name=user.name),
    )


async def _fetch_jwks(issuer: str) -> Dict[str, Any]:
    async with httpx.AsyncClient(timeout=10) as client:
        oidc = await client.get(f"{issuer}/.well-known/openid-configuration")
        oidc.raise_for_status()
        jwks_uri = oidc.json()["jwks_uri"]
        jwks = await client.get(jwks_uri)
        jwks.raise_for_status()
        return jwks.json()


def _get_kid(header: Dict[str, Any]) -> str | None:
    return header.get("kid")


def _verify_id_token(id_token: str, jwks: Dict[str, Any], issuer: str, audience: str) -> Dict[str, Any]:
    headers = jwt.get_unverified_header(id_token)
    kid = _get_kid(headers)
    if not kid:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token header")
    key = next((k for k in jwks.get("keys", []) if k.get("kid") == kid), None)
    if not key:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Signing key not found")
    try:
        claims = jwt.decode(id_token, key, audience=audience, issuer=issuer, options={"verify_at_hash": False})
    except Exception:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid ID token")
    return claims


@router.post("/oidc/verify", response_model=LoginResponse)
async def oidc_verify(payload: OidcVerifyRequest, db: Session = Depends(get_db)):
    if not settings.okta_issuer or not settings.okta_audience:
        raise HTTPException(status_code=500, detail="OIDC not configured")

    jwks = await _fetch_jwks(settings.okta_issuer)
    claims = _verify_id_token(payload.id_token, jwks, settings.okta_issuer, settings.okta_audience)

    email = claims.get("email") or claims.get("preferred_username")
    if not email:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="No email in token")

    user = db.query(User).filter(User.email == email).first()
    if user is None:
        # Do not auto-provision membership; require invite/setup by admin.
        user = User(id=uuid.uuid4(), email=email, password_hash=None, name=claims.get("name"))
        db.add(user)
        db.flush()

    membership = (
        db.query(OrgMembership)
        .filter(OrgMembership.user_id == user.id, OrgMembership.tenant_id == payload.tenantId)
        .first()
    )
    if membership is None:
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="No tenant membership")

    token = create_access_token(
        {
            "userId": str(user.id),
            "tenantId": str(membership.tenant_id),
            "role": membership.role.value,
            "idp": "okta",
        }
    )

    return LoginResponse(
        token=token,
        tenantId=membership.tenant_id,
        role=membership.role,
        user=UserInfo(id=user.id, email=user.email, name=user.name),
    )

