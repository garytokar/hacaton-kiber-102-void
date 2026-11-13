from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload
from app.core.database import get_db
from app.models.models import Fragment, Translation, TranslationVote, User, UserAnswer
from app.schemas.schemas import FragmentOut, TranslationOut
from app.api.deps import get_current_user_optional

router = APIRouter()

def format_fragment_out(frag: Fragment, current_user_id: Optional[int] = None) -> FragmentOut:
    translations_out = []
    for t in frag.translations:
        user_voted = None
        if current_user_id and t.votes:
            for v in t.votes:
                if v.user_id == current_user_id:
                    user_voted = v.vote_type
                    break

        translations_out.append(
            TranslationOut(
                id=t.id,
                fragment_id=t.fragment_id,
                translation_text=t.translation_text,
                status=t.status,
                created_by=t.created_by,
                creator_username=t.creator.username if t.creator else "Система",
                votes_count=t.votes_count,
                reject_count=t.reject_count,
                user_voted=user_voted,
                created_at=t.created_at
            )
        )

    return FragmentOut(
        id=frag.id,
        text_ba=frag.text_ba,
        source_title=frag.source_title,
        source_document=frag.source_document,
        chapter=frag.chapter,
        audio_url=frag.audio_url,
        category=frag.category,
        difficulty=frag.difficulty,
        translations=translations_out
    )

@router.get("/daily", response_model=FragmentOut)
async def get_daily_fragment(
    current_user: Optional[User] = Depends(get_current_user_optional),
    db: AsyncSession = Depends(get_db)
):
    """
    Returns the daily fragment. If user is logged in, selects a fragment they haven't answered yet today,
    or falls back to the day's standard fragment.
    """
    stmt = select(Fragment).options(
        selectinload(Fragment.translations).selectinload(Translation.votes),
        selectinload(Fragment.translations).selectinload(Translation.creator)
    )

    if current_user:
        # Exclude fragments user answered recently
        answered_subq = select(UserAnswer.fragment_id).where(UserAnswer.user_id == current_user.id)
        stmt_unanswered = stmt.where(Fragment.id.not_in(answered_subq)).order_by(Fragment.id)
        res = await db.execute(stmt_unanswered)
        frag = res.scalars().first()
        if frag:
            return format_fragment_out(frag, current_user.id)

    # Fallback to day of year mod total fragments
    res_all = await db.execute(stmt.order_by(Fragment.id))
    all_frags = res_all.scalars().all()
    if not all_frags:
        raise HTTPException(status_code=404, detail="Фрагменты не найдены в базе")

    from datetime import date
    idx = date.today().timetuple().tm_yday % len(all_frags)
    selected = all_frags[idx]
    return format_fragment_out(selected, current_user.id if current_user else None)

@router.get("", response_model=List[FragmentOut])
async def list_fragments(
    category: Optional[str] = None,
    difficulty: Optional[str] = None,
    limit: int = Query(20, le=100),
    offset: int = 0,
    current_user: Optional[User] = Depends(get_current_user_optional),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(Fragment).options(
        selectinload(Fragment.translations).selectinload(Translation.votes),
        selectinload(Fragment.translations).selectinload(Translation.creator)
    )

    if category:
        stmt = stmt.where(Fragment.category == category)
    if difficulty:
        stmt = stmt.where(Fragment.difficulty == difficulty)

    stmt = stmt.order_by(Fragment.id).offset(offset).limit(limit)
    res = await db.execute(stmt)
    frags = res.scalars().all()

    return [format_fragment_out(f, current_user.id if current_user else None) for f in frags]

@router.get("/{fragment_id}", response_model=FragmentOut)
async def get_fragment_by_id(
    fragment_id: int,
    current_user: Optional[User] = Depends(get_current_user_optional),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(Fragment).options(
        selectinload(Fragment.translations).selectinload(Translation.votes),
        selectinload(Fragment.translations).selectinload(Translation.creator)
    ).where(Fragment.id == fragment_id)

    res = await db.execute(stmt)
    frag = res.scalar_one_or_none()
    if not frag:
        raise HTTPException(status_code=404, detail="Фрагмент не найден")

    return format_fragment_out(frag, current_user.id if current_user else None)
