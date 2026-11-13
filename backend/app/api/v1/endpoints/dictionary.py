from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func
from app.core.database import get_db
from app.models.models import DictionaryEntry, User
from app.schemas.schemas import DictionaryEntryCreate, DictionaryEntryOut
from app.api.deps import get_current_user

router = APIRouter()

@router.get("", response_model=List[DictionaryEntryOut])
async def get_user_dictionary(
    search: Optional[str] = None,
    favorite_only: bool = False,
    difficulty: Optional[str] = None,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(DictionaryEntry).where(DictionaryEntry.user_id == user.id)

    if favorite_only:
        stmt = stmt.where(DictionaryEntry.is_favorite == True)
    if difficulty:
        stmt = stmt.where(DictionaryEntry.difficulty == difficulty)
    if search:
        clean_search = search.strip().lower()
        pattern = f"%{clean_search}%"
        stmt = stmt.where(
            (func.lower(DictionaryEntry.word).like(pattern)) |
            (func.lower(DictionaryEntry.translation).like(pattern)) |
            (func.lower(DictionaryEntry.context).like(pattern))
        )

    stmt = stmt.order_by(DictionaryEntry.created_at.desc())
    res = await db.execute(stmt)
    entries = res.scalars().all()

    return [
        DictionaryEntryOut(
            id=e.id,
            word=e.word,
            translation=e.translation,
            context=e.context,
            is_favorite=e.is_favorite,
            difficulty=e.difficulty,
            created_at=e.created_at
        )
        for e in entries
    ]

@router.post("", response_model=DictionaryEntryOut)
async def add_dictionary_entry(
    req: DictionaryEntryCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    entry = DictionaryEntry(
        user_id=user.id,
        fragment_id=req.fragment_id,
        word=req.word.strip(),
        translation=req.translation.strip(),
        context=req.context,
        difficulty=req.difficulty,
        is_favorite=False
    )
    db.add(entry)
    await db.commit()
    await db.refresh(entry)

    return DictionaryEntryOut(
        id=entry.id,
        word=entry.word,
        translation=entry.translation,
        context=entry.context,
        is_favorite=entry.is_favorite,
        difficulty=entry.difficulty,
        created_at=entry.created_at
    )

@router.patch("/{entry_id}/favorite", response_model=DictionaryEntryOut)
async def toggle_favorite(
    entry_id: int,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    res = await db.execute(
        select(DictionaryEntry).where(
            DictionaryEntry.id == entry_id,
            DictionaryEntry.user_id == user.id
        )
    )
    entry = res.scalar_one_or_none()
    if not entry:
        raise HTTPException(status_code=404, detail="Запись словаря не найдена")

    entry.is_favorite = not entry.is_favorite
    await db.commit()
    await db.refresh(entry)

    return DictionaryEntryOut(
        id=entry.id,
        word=entry.word,
        translation=entry.translation,
        context=entry.context,
        is_favorite=entry.is_favorite,
        difficulty=entry.difficulty,
        created_at=entry.created_at
    )

@router.delete("/{entry_id}")
async def delete_dictionary_entry(
    entry_id: int,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    res = await db.execute(
        select(DictionaryEntry).where(
            DictionaryEntry.id == entry_id,
            DictionaryEntry.user_id == user.id
        )
    )
    entry = res.scalar_one_or_none()
    if not entry:
        raise HTTPException(status_code=404, detail="Запись словаря не найдена")

    await db.delete(entry)
    await db.commit()
    return {"success": True, "message": "Слово удалено из словаря"}
