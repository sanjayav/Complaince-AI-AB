 import uuid
 from datetime import datetime
 from fastapi import APIRouter, Depends, HTTPException
 from sqlalchemy.orm import Session

from app.deps import Identity, get_current_identity, get_tenant_scoped_db
 from app.models import Conversation, Message, MessageRole
 from app.schemas import ChatRequest, ChatResponse
from app.rag.retrieval import answer_question_with_retrieval
from app.audit import write_audit


 router = APIRouter()


 @router.post("/", response_model=ChatResponse)
 def chat(
     payload: ChatRequest,
    identity: Identity = Depends(get_current_identity),
    db: Session = Depends(get_tenant_scoped_db),
):
     conversation_id = payload.conversationId
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
         db=db,
         tenant_id=identity.tenant_id,
         question=payload.question,
         filters=payload.filters or {},
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

     return ChatResponse(
         conversationId=conversation_id,
         messageId=assistant_msg.id,
         answer=answer,
         citations=citations,
     )

