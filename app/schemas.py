 import uuid
 from datetime import datetime
 from typing import Any, List, Optional

 from pydantic import BaseModel, EmailStr

from app.models import MessageRole, Role, DocumentType


 # Auth
 class LoginRequest(BaseModel):
     email: EmailStr
     password: str


 class TenantInfo(BaseModel):
     id: uuid.UUID
     name: str


 class UserInfo(BaseModel):
     id: uuid.UUID
     email: EmailStr
     name: Optional[str] = None


 class LoginResponse(BaseModel):
     token: str
     tenantId: uuid.UUID
     role: Role
     user: UserInfo


# OIDC
class OidcVerifyRequest(BaseModel):
    id_token: str
    tenantId: uuid.UUID


 # Conversations
 class ConversationItem(BaseModel):
     id: uuid.UUID
     name: str
     created_at: datetime


 class ConversationsListResponse(BaseModel):
     items: List[ConversationItem]
     next_cursor: Optional[str] = None


 class MessageItem(BaseModel):
     id: uuid.UUID
     role: MessageRole
     content: str
     citations: Optional[Any] = None
     created_at: datetime


 class MessagesListResponse(BaseModel):
     items: List[MessageItem]


 # Chat
 class ChatRequest(BaseModel):
     question: str
     conversationId: Optional[uuid.UUID] = None
     filters: Optional[dict] = None


 class ChatResponse(BaseModel):
     conversationId: uuid.UUID
     messageId: uuid.UUID
     answer: str
     citations: Optional[Any] = None


 # Org
 class AddMemberRequest(BaseModel):
     email: EmailStr
     role: Role


 class MemberItem(BaseModel):
     user_id: uuid.UUID
     email: EmailStr
     name: Optional[str] = None
     role: Role
     created_at: datetime


 class MembersListResponse(BaseModel):
     items: List[MemberItem]


 # Documents
 class IngestRequest(BaseModel):
     s3_key: str
     document_metadata: Optional[dict] = None
    doc_type: Optional[DocumentType] = None
    framework: Optional[str] = None
    jurisdiction: Optional[str] = None
    version: Optional[str] = None


# Compliance Mapping
class MatchItem(BaseModel):
    chunk_id: str
    text: str
    metadata: Optional[Any] = None


class MappingItem(BaseModel):
    source_chunk_id: str
    source_text: str
    matches: List[MatchItem]


class MappingRequest(BaseModel):
    documentId: uuid.UUID
    filters: Optional[dict] = None  # e.g., {"doc_type": "regulation", "framework": "ESRS", ...}
    topK: int = 3
    maxSourceChunks: int = 20


class MappingResponse(BaseModel):
    items: List[MappingItem]


# Compliance Scoring
class ScoreRequest(BaseModel):
    documentId: uuid.UUID
    framework: Optional[str] = None
    jurisdiction: Optional[str] = None
    version: Optional[str] = None


class ScoreResponse(BaseModel):
    score: float  # 0.0 - 1.0
    total_source_chunks: int
    covered_source_chunks: int
    details: Optional[Any] = None


# Frontend /api shapes
class ConversationSummary(BaseModel):
    id: uuid.UUID
    name: str
    createdAt: datetime
    updatedAt: datetime


class ConversationsPageResponse(BaseModel):
    page: int
    limit: int
    totalCount: int
    totalPages: int
    conversations: List[ConversationSummary]


class RenameConversationRequest(BaseModel):
    name: str


class MessageView(BaseModel):
    id: uuid.UUID
    sender: str  # user | ai
    type: str  # text | table | etc.
    content: Any
    metadata: Optional[Any] = None
    createdAt: datetime


class MessagesPageResponse(BaseModel):
    conversationId: uuid.UUID
    page: int
    limit: int
    totalCount: int
    messages: List[MessageView]


class SendMessageRequest(BaseModel):
    conversationId: Optional[uuid.UUID] = None
    question: str


# Documents purge
class PurgeRequest(BaseModel):
    filters: Optional[dict] = None  # {"doc_type"?, "framework"?, "jurisdiction"?, "version"?}
    dryRun: bool = True


class PurgeResponse(BaseModel):
    matched_count: int
    deleted_count: int

