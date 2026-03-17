import os
from datetime import timedelta
from typing import Optional

class Settings:
    app_name: str = os.getenv("APP_NAME", "Compliance Assistant API")
    app_env: str = os.getenv("APP_ENV", "dev")

    database_url: str = os.getenv("DATABASE_URL", "postgresql+psycopg2://postgres:postgres@localhost:5432/compbot")

    jwt_secret: str = os.getenv("JWT_SECRET", "change_me")
    jwt_algorithm: str = os.getenv("JWT_ALGORITHM", "HS256")
    jwt_expires_minutes: int = int(os.getenv("JWT_EXPIRES_MINUTES", "60"))

    aws_region: str = os.getenv("AWS_REGION", "us-east-1")
    s3_bucket: str = os.getenv("S3_BUCKET_NAME", "")
    aws_access_key_id: Optional[str] = os.getenv("AWS_ACCESS_KEY_ID")
    aws_secret_access_key: Optional[str] = os.getenv("AWS_SECRET_ACCESS_KEY")

    openai_api_key: Optional[str] = os.getenv("OPENAI_API_KEY")
    default_model: str = os.getenv("DEFAULT_MODEL", "gpt-4o-mini")
    embeddings_provider: str = os.getenv("EMBEDDINGS_PROVIDER", "openai")

    # CORS / Allowed Origins (comma-separated)
    allowed_origins_raw: str = os.getenv("ALLOWED_ORIGINS", "http://localhost:3000")

    # Okta OIDC
    okta_issuer: Optional[str] = os.getenv("OKTA_ISSUER")
    okta_audience: Optional[str] = os.getenv("OKTA_AUDIENCE")

    # RAG config
    rag_top_k: int = int(os.getenv("RAG_TOP_K", "5"))
    rerank_enabled: bool = os.getenv("RERANK_ENABLED", "true").lower() in ("1", "true", "yes")
    rerank_top_k: int = int(os.getenv("RERANK_TOP_K", "5"))
    hybrid_enabled: bool = os.getenv("HYBRID_ENABLED", "true").lower() in ("1", "true", "yes")
    hybrid_alpha: float = float(os.getenv("HYBRID_ALPHA", "0.7"))  # weight on vector score

    @property
    def jwt_expires_delta(self) -> timedelta:
        return timedelta(minutes=self.jwt_expires_minutes)

    @property
    def allowed_origins(self) -> list[str]:
        return [o.strip() for o in self.allowed_origins_raw.split(",") if o.strip()]

    @property
    def is_prod(self) -> bool:
        return self.app_env.lower() in ("prod", "production")


settings = Settings()
