from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from app.core.database import get_db
from app.models.models import DailyQuest, UserQuest, WeeklyExpedition, User
from app.schemas.schemas import DailyQuestOut, WeeklyExpeditionOut, TaskSubmissionResult
from app.api.deps import get_current_user, get_current_user_optional
from app.services.gamification import add_xp_and_currency

router = APIRouter()

@router.get("/daily", response_model=List[DailyQuestOut])
async def get_daily_quests(
    user: Optional[User] = Depends(get_current_user_optional),
    db: AsyncSession = Depends(get_db)
):
    today = date.today()
    all_quests_res = await db.execute(select(DailyQuest))
    quests = all_quests_res.scalars().all()

    out = []
    for q in quests:
        progress = 0
        completed = False
        claimed = False
        if user:
            uq_res = await db.execute(
                select(UserQuest).where(
                    UserQuest.user_id == user.id,
                    UserQuest.quest_id == q.id,
                    UserQuest.quest_date == today
                )
            )
            uq = uq_res.scalar_one_or_none()
            if uq:
                progress = uq.progress
                completed = uq.completed
                claimed = uq.claimed

        out.append(
            DailyQuestOut(
                id=q.id,
                code=q.code,
                name=q.name,
                description=q.description,
                target=q.target,
                reward_xp=q.reward_xp,
                reward_currency=q.reward_currency,
                progress=progress,
                completed=completed,
                claimed=claimed
            )
        )
    return out

@router.post("/{quest_id}/claim", response_model=TaskSubmissionResult)
async def claim_quest_reward(
    quest_id: int,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    today = date.today()
    q_res = await db.execute(select(DailyQuest).where(DailyQuest.id == quest_id))
    quest = q_res.scalar_one_or_none()
    if not quest:
        raise HTTPException(status_code=404, detail="Квест не найден")

    uq_res = await db.execute(
        select(UserQuest).where(
            UserQuest.user_id == user.id,
            UserQuest.quest_id == quest_id,
            UserQuest.quest_date == today
        )
    )
    uq = uq_res.scalar_one_or_none()
    if not uq or not uq.completed:
        raise HTTPException(status_code=400, detail="Квест ещё не завершён")
    if uq.claimed:
        raise HTTPException(status_code=400, detail="Награда за этот квест уже получена")

    uq.claimed = True
    stats, level_up, new_level = await add_xp_and_currency(db, user.id, quest.reward_xp, quest.reward_currency)
    await db.commit()

    return TaskSubmissionResult(
        success=True,
        xp_awarded=quest.reward_xp,
        currency_awarded=quest.reward_currency,
        new_level=new_level if level_up else None,
        level_up=level_up,
        streak_updated=False,
        current_streak=stats.current_streak,
        message=f"Награда за квест «{quest.name}» получена!"
    )

@router.get("/expedition", response_model=WeeklyExpeditionOut)
async def get_weekly_expedition(db: AsyncSession = Depends(get_db)):
    res = await db.execute(
        select(WeeklyExpedition).where(WeeklyExpedition.is_active == True).limit(1)
    )
    exp = res.scalar_one_or_none()
    if not exp:
        raise HTTPException(status_code=404, detail="Активная экспедиция не найдена")

    pct = round(min(100.0, (exp.current_count / exp.target_count) * 100), 1) if exp.target_count > 0 else 0.0

    return WeeklyExpeditionOut(
        id=exp.id,
        title=exp.title,
        description=exp.description,
        target_count=exp.target_count,
        current_count=exp.current_count,
        reward_xp=exp.reward_xp,
        reward_currency=exp.reward_currency,
        progress_percentage=pct
    )
