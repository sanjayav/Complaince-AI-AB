 import uuid
 from typing import Optional

 from fastapi import APIRouter, Depends, Query
 from sqlalchemy import and_, desc
 from sqlalchemy.orm import Session

from app.deps import Identity, get_current_identity, get_tenant_scoped_db
 from app.models import Conversation, Message
 from app.schemas import ConversationItem, ConversationsListResponse, MessageItem, MessagesListResponse


 router = APIRouter()


 @router.get("/", response_model=ConversationsListResponse)
 def list_conversations(
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
     cursor: Optional[int] = Query(default=0, ge=0),
     limit: int = Query(default=20, ge=1, le=100),
):
     q = (
         db.query(Conversation)
         .filter(and_(Conversation.tenant_id == identity.tenant_id, Conversation.deleted_at.is_(None)))
         .order_by(desc(Conversation.created_at), desc(Conversation.id))
         .offset(cursor)
         .limit(limit + 1)
     )
     rows = q.all()
     items = [
         ConversationItem(id=r.id, name=r.name, created_at=r.created_at)
         for r in rows[:limit]
     ]
     next_cursor = cursor + limit if len(rows) > limit else None
     return ConversationsListResponse(items=items, next_cursor=str(next_cursor) if next_cursor is not None else None)


 @router.get("/{conversation_id}/messages", response_model=MessagesListResponse)
 def get_messages(
     conversation_id: uuid.UUID,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
     conv = (
         db.query(Conversation)
         .filter(Conversation.id == conversation_id, Conversation.tenant_id == identity.tenant_id)
         .first()
     )
     if not conv:
         return MessagesListResponse(items=[])

     msgs = (
         db.query(Message)
         .filter(Message.conversation_id == conversation_id, Message.tenant_id == identity.tenant_id)
         .order_by(Message.created_at.asc(), Message.id.asc())
         .all()
     )
     items = [
         MessageItem(
             id=m.id,
             role=m.role,
             content=m.content,
             citations=m.citations,
             created_at=m.created_at,
         )
         for m in msgs
     ]
     return MessagesListResponse(items=items)

