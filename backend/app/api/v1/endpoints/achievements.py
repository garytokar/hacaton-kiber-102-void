from typing import List, Optional
from fastapi import APIRouter, Depends
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from app.core.database import get_db
from app.models.models import Achievement, UserAchievement, User
from app.schemas.schemas import AchievementOut
from app.api.deps import get_current_user_optional

router = APIRouter()

@router.get("", response_model=List[AchievementOut])
async def list_achievements(
    user: Optional[User] = Depends(get_current_user_optional),
    db: AsyncSession = Depends(get_db)
):
    ach_res = await db.execute(select(Achievement).order_by(Achievement.id))
    achievements = ach_res.scalars().all()

    unlocked_map = {}
    if user:
        ua_res = await db.execute(
            select(UserAchievement).where(UserAchievement.user_id == user.id)
        )
        for ua in ua_res.scalars().all():
            unlocked_map[ua.achievement_id] = ua.unlocked_at

    return [
        AchievementOut(
            id=a.id,
            code=a.code,
            name=a.name,
            description=a.description,
            icon=a.icon,
            reward_xp=a.reward_xp,
            reward_currency=a.reward_currency,
            unlocked=a.id in unlocked_map,
            unlocked_at=unlocked_map.get(a.id)
        )
        for a in achievements
    ]
