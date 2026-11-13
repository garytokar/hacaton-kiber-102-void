from datetime import date
from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from app.core.database import get_db
from app.models.models import (
    Translation, TranslationVote, TranslationHistory, Fragment, User, UserAnswer, UserStats
)
from app.schemas.schemas import (
    TranslationOut, TranslationCreate, TranslationVoteRequest,
    TranslationHistoryOut, TaskSubmissionResult
)
from app.api.deps import get_current_user
from app.services.gamification import (
    add_xp_and_currency, update_user_streak, update_quest_progress, check_achievements
)
from app.services.consensus import find_matching_translation
from app.services.moderation import evaluate_translation_status

router = APIRouter()

@router.get("/{fragment_id}", response_model=List[TranslationOut])
async def get_fragment_translations(
    fragment_id: int,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(Translation).options(
        selectinload(Translation.votes),
        selectinload(Translation.creator)
    ).where(Translation.fragment_id == fragment_id)
    res = await db.execute(stmt)
    translations = res.scalars().all()

    out = []
    for t in translations:
        user_voted = None
        for v in t.votes:
            if v.user_id == user.id:
                user_voted = v.vote_type
                break
        out.append(
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
    return out

@router.post("", response_model=TaskSubmissionResult)
async def propose_translation(
    req: TranslationCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    # Verify fragment exists
    frag_res = await db.execute(select(Fragment).where(Fragment.id == req.fragment_id))
    fragment = frag_res.scalar_one_or_none()
    if not fragment:
        raise HTTPException(status_code=404, detail="Фрагмент не найден")

    clean_text = req.translation_text.strip()
    if not clean_text or len(clean_text) < 3:
        raise HTTPException(status_code=400, detail="Текст перевода слишком короткий")

    # Fuzzy consensus check (ported from Go fuzzy pkg)
    existing_res = await db.execute(
        select(Translation).where(Translation.fragment_id == req.fragment_id)
    )
    existing_trans = existing_res.scalars().all()
    fuzzy_match = find_matching_translation(clean_text, existing_trans, threshold=0.82)

    if fuzzy_match:
        target_trans, sim_score = fuzzy_match
        target_trans_id = target_trans.id
        msg_note = f"Перевод признан идентичным существующему (сходство {int(sim_score*100)}%) — зачтён голос консенсуса!"

        # Check if user already voted on this matched translation
        vote_check = await db.execute(
            select(TranslationVote).where(
                TranslationVote.translation_id == target_trans.id,
                TranslationVote.user_id == user.id
            )
        )
        if not vote_check.scalar_one_or_none():
            user_vote = TranslationVote(
                translation_id=target_trans.id,
                user_id=user.id,
                vote_type="confirm"
            )
            db.add(user_vote)
            target_trans.votes_count += 1
            if target_trans.votes_count >= 3:
                target_trans.status = "confirmed"

        # Record UserAnswer
        answer = UserAnswer(
            user_id=user.id,
            fragment_id=fragment.id,
            translation_id=target_trans.id,
            action="propose"
        )
        db.add(answer)
    else:
        # Create brand new translation
        new_trans = Translation(
            fragment_id=req.fragment_id,
            translation_text=clean_text,
            status="pending",
            created_by=user.id,
            votes_count=1,
            reject_count=0
        )
        db.add(new_trans)
        await db.flush()
        target_trans_id = new_trans.id
        msg_note = "Ваш перевод сохранён и опубликован для голосования сообщества!"

        # User automatically votes confirm on their proposal
        user_vote = TranslationVote(
            translation_id=new_trans.id,
            user_id=user.id,
            vote_type="confirm"
        )
        db.add(user_vote)

        # History
        history = TranslationHistory(
            translation_id=new_trans.id,
            previous_text="",
            new_text=clean_text,
            changed_by=user.id
        )
        db.add(history)

        # Record UserAnswer
        answer = UserAnswer(
            user_id=user.id,
            fragment_id=fragment.id,
            translation_id=new_trans.id,
            action="propose"
        )
        db.add(answer)

    # Gamification
    stats_res = await db.execute(select(UserStats).where(UserStats.user_id == user.id))
    stats = stats_res.scalar_one_or_none()
    if not stats:
        stats = UserStats(user_id=user.id)
        db.add(stats)
        await db.flush()

    stats.total_proposals_made += 1
    stats.total_tasks_completed += 1

    today = date.today()
    streak_updated, cur_streak = await update_user_streak(stats, today)

    # XP for proposing new translation: +8 XP, +5 currency
    base_xp = 8
    currency_award = 5
    stats, level_up, new_level = await add_xp_and_currency(db, user.id, base_xp, currency_award)

    # Daily quest updates
    await update_quest_progress(db, user.id, "propose", 1)
    await update_quest_progress(db, user.id, "verify", 1)

    # Check achievements
    new_ach = await check_achievements(db, user.id)

    await db.commit()

    return TaskSubmissionResult(
        success=True,
        xp_awarded=base_xp,
        currency_awarded=currency_award,
        new_level=new_level if level_up else None,
        level_up=level_up,
        streak_updated=streak_updated,
        current_streak=cur_streak,
        achievement_unlocked=new_ach,
        message=msg_note
    )

@router.post("/{translation_id}/vote", response_model=TaskSubmissionResult)
async def vote_translation(
    translation_id: int,
    req: TranslationVoteRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(Translation).options(selectinload(Translation.fragment)).where(Translation.id == translation_id)
    res = await db.execute(stmt)
    translation = res.scalar_one_or_none()
    if not translation:
        raise HTTPException(status_code=404, detail="Перевод не найден")

    # Check if already voted
    existing_vote_res = await db.execute(
        select(TranslationVote).where(
            TranslationVote.translation_id == translation_id,
            TranslationVote.user_id == user.id
        )
    )
    if existing_vote_res.scalar_one_or_none():
        raise HTTPException(status_code=400, detail="Вы уже проголосовали за этот вариант перевода")

    # Record vote
    vote = TranslationVote(
        translation_id=translation_id,
        user_id=user.id,
        vote_type=req.vote_type
    )
    db.add(vote)
    await db.flush()

    # Re-evaluate consensus
    new_status = await evaluate_translation_status(db, translation_id, user.id)

    # Record user answer
    answer = UserAnswer(
        user_id=user.id,
        fragment_id=translation.fragment_id,
        translation_id=translation.id,
        action=req.vote_type
    )
    db.add(answer)

    # Gamification
    stats_res = await db.execute(select(UserStats).where(UserStats.user_id == user.id))
    stats = stats_res.scalar_one_or_none()
    if not stats:
        stats = UserStats(user_id=user.id)
        db.add(stats)
        await db.flush()

    stats.total_tasks_completed += 1
    streak_updated = False
    cur_streak = stats.current_streak

    if req.vote_type == "confirm":
        stats.total_translations_confirmed += 1
        today = date.today()
        streak_updated, cur_streak = await update_user_streak(stats, today)

    # Base XP calculation
    # Hard fragments yield +8 XP, normal +5 XP
    is_hard = translation.fragment and translation.fragment.difficulty == "hard"
    xp_award = 8 if is_hard else 5
    curr_award = 2
    stats, level_up, new_level = await add_xp_and_currency(db, user.id, xp_award, curr_award)

    # Update quest progress
    await update_quest_progress(db, user.id, "confirm" if req.vote_type == "confirm" else "verify", 1)
    await update_quest_progress(db, user.id, "verify", 1)

    # Check achievements
    new_ach = await check_achievements(db, user.id)

    await db.commit()

    msg = "Голос подтверждения учтён! Спасибо за вклад в сохранение языка." if req.vote_type == "confirm" else "Замечание учтено и передано на экспертную проверку."

    return TaskSubmissionResult(
        success=True,
        xp_awarded=xp_award,
        currency_awarded=curr_award,
        new_level=new_level if level_up else None,
        level_up=level_up,
        streak_updated=streak_updated,
        current_streak=cur_streak,
        achievement_unlocked=new_ach,
        message=msg
    )

@router.get("/{translation_id}/history", response_model=List[TranslationHistoryOut])
async def get_translation_history(
    translation_id: int,
    db: AsyncSession = Depends(get_db)
):
    stmt = (
        select(TranslationHistory, User)
        .outerjoin(User, TranslationHistory.changed_by == User.id)
        .where(TranslationHistory.translation_id == translation_id)
        .order_by(TranslationHistory.created_at.desc())
    )
    res = await db.execute(stmt)
    records = res.all()

    return [
        TranslationHistoryOut(
            id=hist.id,
            translation_id=hist.translation_id,
            previous_text=hist.previous_text,
            new_text=hist.new_text,
            changed_by_username=(u.full_name or u.username) if u else "Хранитель языка",
            created_at=hist.created_at
        )
        for hist, u in records
    ]
