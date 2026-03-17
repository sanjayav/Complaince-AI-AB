 from datetime import datetime, timedelta, timezone
 from typing import Any, Dict

 from jose import jwt
 from passlib.context import CryptContext

 from app.config import settings


 password_ctx = CryptContext(schemes=["bcrypt"], deprecated="auto")


 def hash_password(plain_password: str) -> str:
     return password_ctx.hash(plain_password)


 def verify_password(plain_password: str, password_hash: str) -> bool:
     return password_ctx.verify(plain_password, password_hash)


 def create_access_token(subject: Dict[str, Any]) -> str:
     to_encode = subject.copy()
     expire = datetime.now(timezone.utc) + settings.jwt_expires_delta
     to_encode.update({"exp": expire})
     token = jwt.encode(to_encode, settings.jwt_secret, algorithm=settings.jwt_algorithm)
     return token


 def decode_access_token(token: str) -> Dict[str, Any]:
     return jwt.decode(token, settings.jwt_secret, algorithms=[settings.jwt_algorithm])

