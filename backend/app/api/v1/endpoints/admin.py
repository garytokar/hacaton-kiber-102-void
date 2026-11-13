from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func, desc
from sqlalchemy.orm import selectinload
from app.core.database import get_db
from app.models.models import (
    User, Fragment, Translation, TranslationHistory, School, SchoolClass, UserStats, UserAnswer, DictionaryEntry
)
from pydantic import BaseModel
from app.schemas.schemas import (
    UserOut, AdminReviewItem, FragmentCreateAdmin, SystemOverviewStats
)
from app.api.deps import get_current_admin, get_current_expert_or_admin
from app.api.v1.endpoints.auth import format_user_out

router = APIRouter()

@router.get("/users", response_model=List[UserOut])
async def list_admin_users(
    search: Optional[str] = None,
    role: Optional[str] = None,
    limit: int = Query(50, le=100),
    offset: int = 0,
    admin: User = Depends(get_current_admin),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(User).options(
        selectinload(User.stats),
        selectinload(User.school),
        selectinload(User.school_class)
    )

    if role:
        stmt = stmt.where(User.role == role)
    if search:
        p = f"%{search.lower()}%"
        stmt = stmt.where((User.username.ilike(p)) | (User.email.ilike(p)) | (User.full_name.ilike(p)))

    stmt = stmt.order_by(User.id).offset(offset).limit(limit)
    res = await db.execute(stmt)
    users = res.scalars().all()

    return [format_user_out(u) for u in users]

@router.patch("/users/{user_id}/role", response_model=UserOut)
async def update_user_role(
    user_id: int,
    role: str = Query(..., pattern="^(student|teacher|expert|admin)$"),
    admin: User = Depends(get_current_admin),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(User).options(
        selectinload(User.stats),
        selectinload(User.school),
        selectinload(User.school_class)
    ).where(User.id == user_id)
    res = await db.execute(stmt)
    user = res.scalar_one_or_none()
    if not user:
        raise HTTPException(status_code=404, detail="Пользователь не найден")

    user.role = role
    await db.commit()
    await db.refresh(user)

    return format_user_out(user)

@router.get("/review", response_model=List[AdminReviewItem])
async def get_translations_for_review(
    status: Optional[str] = Query(None, pattern="^(pending|expert_review|rejected)$"),
    expert: User = Depends(get_current_expert_or_admin),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(Translation).join(Translation.fragment).options(
        selectinload(Translation.fragment),
        selectinload(Translation.creator)
    )

    if status:
        stmt = stmt.where(Translation.status == status)
    else:
        stmt = stmt.where(Translation.status.in_(["pending", "expert_review"]))

    stmt = stmt.order_by(desc(Translation.created_at))
    res = await db.execute(stmt)
    translations = res.scalars().all()

    out = []
    for t in translations:
        out.append(
            AdminReviewItem(
                id=t.id,
                fragment_id=t.fragment_id,
                text_ba=t.fragment.text_ba,
                source_title=t.fragment.source_title,
                translation_id=t.id,
                translation_text=t.translation_text,
                status=t.status,
                votes_count=t.votes_count,
                reject_count=t.reject_count,
                creator_username=t.creator.username if t.creator else "Система",
                created_at=t.created_at
            )
        )
    return out

@router.post("/review/{translation_id}/approve")
async def approve_translation(
    translation_id: int,
    expert: User = Depends(get_current_expert_or_admin),
    db: AsyncSession = Depends(get_db)
):
    res = await db.execute(select(Translation).where(Translation.id == translation_id))
    translation = res.scalar_one_or_none()
    if not translation:
        raise HTTPException(status_code=404, detail="Перевод не найден")

    old_status = translation.status
    translation.status = "confirmed"

    # Log history
    db.add(TranslationHistory(
        translation_id=translation.id,
        previous_text=f"Статус: {old_status}",
        new_text=f"Подтверждено экспертом/модератором {expert.username}",
        changed_by=expert.id
    ))

    # Reward creator if accepted proposal
    if translation.created_by:
        creator_stats_res = await db.execute(
            select(UserStats).where(UserStats.user_id == translation.created_by)
        )
        creator_stats = creator_stats_res.scalar_one_or_none()
        if creator_stats:
            creator_stats.total_accepted_proposals += 1
            creator_stats.xp += 15
            creator_stats.currency += 10

    await db.commit()
    return {"success": True, "message": "Перевод успешно подтверждён и добавлен в Сокровищницу!"}

@router.post("/review/{translation_id}/reject")
async def reject_translation(
    translation_id: int,
    reason: Optional[str] = None,
    expert: User = Depends(get_current_expert_or_admin),
    db: AsyncSession = Depends(get_db)
):
    res = await db.execute(select(Translation).where(Translation.id == translation_id))
    translation = res.scalar_one_or_none()
    if not translation:
        raise HTTPException(status_code=404, detail="Перевод не найден")

    old_status = translation.status
    translation.status = "rejected"

    db.add(TranslationHistory(
        translation_id=translation.id,
        previous_text=f"Статус: {old_status}",
        new_text=f"Отклонено модератором {expert.username}. Причина: {reason or 'Несоответствие нормам языка'}",
        changed_by=expert.id
    ))

    await db.commit()
    return {"success": True, "message": "Перевод отклонён"}

@router.post("/fragments")
async def add_fragment_admin(
    req: FragmentCreateAdmin,
    expert: User = Depends(get_current_expert_or_admin),
    db: AsyncSession = Depends(get_db)
):
    frag = Fragment(
        text_ba=req.text_ba.strip(),
        source_title=req.source_title.strip(),
        source_document=req.source_document or "Корпус башкирского языка",
        category=req.category,
        difficulty=req.difficulty
    )
    db.add(frag)
    await db.flush()

    if req.initial_translation and req.initial_translation.strip():
        trans = Translation(
            fragment_id=frag.id,
            language_pair_id=1,
            translation_text=req.initial_translation.strip(),
            status="pending",
            created_by=expert.id,
            votes_count=1,
            reject_count=0
        )
        db.add(trans)

    await db.commit()
    return {"success": True, "id": frag.id, "message": "Фрагмент успешно добавлен в учебную базу!"}

@router.get("/fragments")
async def list_admin_fragments(
    limit: int = 50,
    offset: int = 0,
    category: Optional[str] = None,
    expert: User = Depends(get_current_expert_or_admin),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(Fragment).options(
        selectinload(Fragment.translations)
    ).order_by(desc(Fragment.id)).offset(offset).limit(limit)
    if category:
        stmt = stmt.where(Fragment.category == category)
    res = await db.execute(stmt)
    frags = res.scalars().all()
    return [
        {
            "id": f.id,
            "text_ba": f.text_ba,
            "source_title": f.source_title,
            "category": f.category,
            "difficulty": f.difficulty,
            "translations_count": len(f.translations),
            "confirmed_count": len([t for t in f.translations if t.status == "confirmed"])
        }
        for f in frags
    ]

@router.delete("/fragments/{fragment_id}")
async def delete_fragment_admin(
    fragment_id: int,
    expert: User = Depends(get_current_expert_or_admin),
    db: AsyncSession = Depends(get_db)
):
    frag = (await db.execute(select(Fragment).where(Fragment.id == fragment_id))).scalar_one_or_none()
    if not frag:
        raise HTTPException(status_code=404, detail="Фрагмент не найден")
    await db.delete(frag)
    await db.commit()
    return {"success": True, "message": "Фрагмент успешно удалён из корпуса"}

class SchoolCreateAdmin(BaseModel):
    name: str
    city: str
    class_names: List[str] = []

@router.post("/schools")
async def create_school_admin(
    req: SchoolCreateAdmin,
    admin: User = Depends(get_current_admin),
    db: AsyncSession = Depends(get_db)
):
    school = School(name=req.name.strip(), city=req.city.strip())
    db.add(school)
    await db.flush()

    for c_name in req.class_names:
        if c_name.strip():
            db.add(SchoolClass(school_id=school.id, name=c_name.strip()))

    await db.commit()
    return {"success": True, "message": f"Школа «{school.name}» успешно добавлена!"}

class AdminDictionaryCreate(BaseModel):
    word: str
    translation: str
    context: Optional[str] = None
    difficulty: str = "medium"
    broadcast_to_all_students: bool = True

@router.post("/dictionary")
async def add_dictionary_word_admin(
    req: AdminDictionaryCreate,
    expert: User = Depends(get_current_expert_or_admin),
    db: AsyncSession = Depends(get_db)
):
    """Add a dictionary entry and distribute it to all students."""
    word = req.word.strip()
    translation = req.translation.strip()
    context = req.context.strip() if req.context else None

    if req.broadcast_to_all_students:
        user_ids_res = await db.execute(select(User.id))
        target_ids = user_ids_res.scalars().all()
    else:
        target_ids = [expert.id]

    added_count = 0
    for uid in target_ids:
        exists = (await db.execute(
            select(DictionaryEntry).where(
                DictionaryEntry.user_id == uid,
                DictionaryEntry.word == word
            )
        )).scalar_one_or_none()

        if not exists:
            db.add(DictionaryEntry(
                user_id=uid,
                word=word,
                translation=translation,
                context=context,
                difficulty=req.difficulty,
                is_favorite=False
            ))
            added_count += 1

    await db.commit()
    return {
        "success": True,
        "message": f"Слово «{word}» успешно добавлено в словарь {added_count} пользователей!",
        "added_count": added_count
    }

@router.get("/stats", response_model=SystemOverviewStats)
async def get_admin_system_stats(
    expert: User = Depends(get_current_expert_or_admin),
    db: AsyncSession = Depends(get_db)
):
    u_cnt = (await db.execute(select(func.count(User.id)))).scalar() or 0
    f_cnt = (await db.execute(select(func.count(Fragment.id)))).scalar() or 0
    t_cnt = (await db.execute(select(func.count(Translation.id)))).scalar() or 0
    c_cnt = (await db.execute(select(func.count(Translation.id)).where(Translation.status == "confirmed"))).scalar() or 0
    r_cnt = (await db.execute(select(func.count(Translation.id)).where(Translation.status.in_(["pending", "expert_review"])))).scalar() or 0
    ans_cnt = (await db.execute(select(func.count(UserAnswer.id)))).scalar() or 0
    sch_cnt = (await db.execute(select(func.count(School.id)))).scalar() or 0

    return SystemOverviewStats(
        total_users=u_cnt,
        total_fragments=f_cnt,
        total_translations=t_cnt,
        confirmed_translations=c_cnt,
        in_review_translations=r_cnt,
        total_contributions=ans_cnt,
        active_schools_count=sch_cnt
    )
