from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import time

from app.config import settings
from app.routers.auth import router as auth_router
from app.routers.conversations import router as conversations_router
from app.routers.chat import router as chat_router
from app.routers.documents import router as documents_router
from app.routers.org import router as org_router
from app.routers.compliance import router as compliance_router
from app.routers.api_frontend import router as api_router


app = FastAPI(title=settings.app_name)

# CORS: allow configured origins; if wildcard, disable credentials
allow_origins = settings.allowed_origins or ["http://localhost:3000"]
wildcard = any(o == "*" for o in allow_origins)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"] if wildcard else allow_origins,
    allow_credentials=False if wildcard else True,
    allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"],
    allow_headers=["*"],
)


@app.get("/health")
def health():
    return {"status": "ok"}


app.include_router(auth_router, prefix="/auth", tags=["auth"])
app.include_router(conversations_router, prefix="/conversations", tags=["conversations"])
app.include_router(chat_router, prefix="/chat", tags=["chat"])
app.include_router(documents_router, prefix="/documents", tags=["documents"])
app.include_router(org_router, prefix="/org", tags=["org"])
app.include_router(compliance_router, prefix="/compliance", tags=["compliance"])
app.include_router(api_router, prefix="/api", tags=["api"])
