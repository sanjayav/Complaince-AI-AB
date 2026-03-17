import uuid
from datetime import datetime
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy import and_, desc, asc, func
from sqlalchemy.orm import Session

from app.deps import Identity, get_current_identity, get_tenant_scoped_db
from app.models import Conversation, Message, MessageRole
from app.schemas import (
    ConversationsPageResponse,
    ConversationSummary,
    MessagesPageResponse,
    MessageView,
    RenameConversationRequest,
    SendMessageRequest,
)
from app.rag.retrieval import answer_question_with_retrieval
from app.audit import write_audit


router = APIRouter()


def _conv_updated_at(db: Session, conv_id: uuid.UUID, fallback: datetime) -> datetime:
    latest = (
        db.query(func.max(Message.created_at))
        .filter(Message.conversation_id == conv_id)
        .scalar()
    )
    return latest or fallback


@router.get("/conversations", response_model=ConversationsPageResponse)
def list_conversations_api(
    page: int = Query(default=1, ge=1),
    limit: int = Query(default=20, ge=1, le=100),
    sort: str = Query(default="desc", pattern="^(asc|desc)$"),
    search: Optional[str] = Query(default=None),
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
    base = db.query(Conversation).filter(
        and_(Conversation.tenant_id == identity.tenant_id, Conversation.deleted_at.is_(None))
    )
    if search:
        base = base.filter(Conversation.name.ilike(f"%{search}%"))

    total = base.count()
    total_pages = (total + limit - 1) // limit
    order_clause = asc(Conversation.created_at) if sort == "asc" else desc(Conversation.created_at)
    rows = base.order_by(order_clause, Conversation.id).offset((page - 1) * limit).limit(limit).all()

    items = [
        ConversationSummary(
            id=r.id,
            name=r.name,
            createdAt=r.created_at,
            updatedAt=_conv_updated_at(db, r.id, r.created_at),
        )
        for r in rows
    ]
    return ConversationsPageResponse(
        page=page, limit=limit, totalCount=total, totalPages=total_pages, conversations=items
    )


@router.get("/conversations/{conversation_id}/messages", response_model=MessagesPageResponse)
def list_messages_api(
    conversation_id: uuid.UUID,
    page: int = Query(default=1, ge=1),
    limit: int = Query(default=20, ge=1, le=200),
    sort: str = Query(default="asc", pattern="^(asc|desc)$"),
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
    conv = (
        db.query(Conversation)
        .filter(Conversation.id == conversation_id, Conversation.tenant_id == identity.tenant_id)
        .first()
    )
    if not conv:
        raise HTTPException(status_code=404, detail="Conversation not found")

    base = db.query(Message).filter(
        Message.conversation_id == conversation_id, Message.tenant_id == identity.tenant_id
    )
    total = base.count()
    total_pages = (total + limit - 1) // limit
    order_clause = asc(Message.created_at) if sort == "asc" else desc(Message.created_at)
    rows = base.order_by(order_clause, Message.id).offset((page - 1) * limit).limit(limit).all()

    def _sender(role: MessageRole) -> str:
        return "ai" if role == MessageRole.assistant else "user"

    items = [
        MessageView(
            id=m.id,
            sender=_sender(m.role),
            type="text",
            content=m.content,
            metadata=m.citations if m.citations is not None else None,
            createdAt=m.created_at,
        )
        for m in rows
    ]
    return MessagesPageResponse(
        conversationId=conversation_id, page=page, limit=limit, totalCount=total, messages=items
    )


@router.patch("/conversations/{conversation_id}")
def rename_conversation_api(
    conversation_id: uuid.UUID,
    payload: RenameConversationRequest,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
    conv = (
        db.query(Conversation)
        .filter(Conversation.id == conversation_id, Conversation.tenant_id == identity.tenant_id)
        .first()
    )
    if not conv:
        raise HTTPException(status_code=404, detail="Conversation not found")
    conv.name = payload.name
    db.add(conv)
    write_audit(
        db=db,
        tenant_id=identity.tenant_id,
        user_id=identity.user_id,
        action="conversation.renamed",
        entity_type="conversation",
        entity_id=str(conversation_id),
        metadata={"name": payload.name},
    )
    db.commit()
    return {
        "id": str(conv.id),
        "name": conv.name,
        "updatedAt": datetime.utcnow().isoformat() + "Z",
    }


@router.delete("/conversations/{conversation_id}")
def delete_conversation_api(
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
        return {"message": "Conversation deleted successfully.", "deletedId": str(conversation_id)}
    db.delete(conv)
    write_audit(
        db=db,
        tenant_id=identity.tenant_id,
        user_id=identity.user_id,
        action="conversation.deleted",
        entity_type="conversation",
        entity_id=str(conversation_id),
        metadata=None,
    )
    db.commit()
    return {"message": "Conversation deleted successfully.", "deletedId": str(conversation_id)}


@router.post("/messages")
def send_message_api(
    payload: SendMessageRequest,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
    conversation_id = payload.conversationId
    created_new = False
    if conversation_id is None:
        conv = Conversation(
            id=uuid.uuid4(),
            tenant_id=identity.tenant_id,
            user_id=identity.user_id,
            name=payload.question[:60] or "Conversation",
            created_at=datetime.utcnow(),
        )
        db.add(conv)
        db.flush()
        conversation_id = conv.id
        created_new = True

    user_msg = Message(
        id=uuid.uuid4(),
        tenant_id=identity.tenant_id,
        conversation_id=conversation_id,
        user_id=identity.user_id,
        role=MessageRole.user,
        content=payload.question,
        created_at=datetime.utcnow(),
    )
    db.add(user_msg)
    db.flush()

    answer, citations = answer_question_with_retrieval(
        db=db, tenant_id=identity.tenant_id, question=payload.question, filters={}
    )

    assistant_msg = Message(
        id=uuid.uuid4(),
        tenant_id=identity.tenant_id,
        conversation_id=conversation_id,
        user_id=None,
        role=MessageRole.assistant,
        content=answer,
        citations=citations,
        created_at=datetime.utcnow(),
    )
    db.add(assistant_msg)
    write_audit(
        db=db,
        tenant_id=identity.tenant_id,
        user_id=identity.user_id,
        action="chat.answer",
        entity_type="conversation",
        entity_id=str(conversation_id),
        metadata={"question_len": len(payload.question), "answer_len": len(answer)},
    )
    db.commit()

    message_obj = {
        "id": str(assistant_msg.id),
        "sender": "ai",
        "type": "text",
        "content": assistant_msg.content,
        "createdAt": assistant_msg.created_at.isoformat() + "Z",
    }
    if created_new:
        return {
            "conversation": {
                "id": str(conversation_id),
                "name": payload.question[:60] or "Conversation",
                "createdAt": user_msg.created_at.isoformat() + "Z",
            },
            "message": message_obj,
        }
    else:
        return {
            "conversationId": str(conversation_id),
            "message": message_obj,
        }


