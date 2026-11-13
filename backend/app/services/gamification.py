from datetime import datetime, date, timedelta
from typing import Optional, Tuple, List
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, update
from app.models.models import (
    User, UserStats, Achievement, UserAchievement, DailyQuest, UserQuest, WeeklyExpedition
)
from app.schemas.schemas import AchievementOut

LEVEL_THRESHOLDS = [
    (1, "Начинающий", 0),
    (2, "Ученик", 100),
    (3, "Искатель слов", 250),
    (4, "Знаток", 500),
    (5, "Переводчик", 900),
    (6, "Хранитель", 1500),
    (7, "Мастер языка", 2500),
    (8, "Наставник", 4000),
    (9, "Мудрец", 6500),
    (10, "Легенда языка", 10000),
]

def get_level_info(xp: int) -> Tuple[int, str, int]:
    """Returns (current_level, level_title, next_level_xp)"""
    current_level = 1
    current_title = LEVEL_THRESHOLDS[0][1]
    next_level_xp = LEVEL_THRESHOLDS[1][2]

    for i in range(len(LEVEL_THRESHOLDS)):
        lvl, title, req_xp = LEVEL_THRESHOLDS[i]
        if xp >= req_xp:
            current_level = lvl
            current_title = title
            if i + 1 < len(LEVEL_THRESHOLDS):
                next_level_xp = LEVEL_THRESHOLDS[i + 1][2]
            else:
                next_level_xp = LEVEL_THRESHOLDS[-1][2]  # max level
        else:
            break
            
    return current_level, current_title, next_level_xp

async def update_user_streak(stats: UserStats, today: date) -> Tuple[bool, int]:
    """Updates streak. Returns (streak_incremented, current_streak)"""
    if stats.last_activity_date == today:
        return False, stats.current_streak
    
    incremented = False
    if stats.last_activity_date == today - timedelta(days=1):
        stats.current_streak += 1
        incremented = True
    elif stats.last_activity_date is None:
        stats.current_streak = 1
        incremented = True
    else:
        # Missed at least one day
        if stats.streak_protected:
            # Protected by item
            stats.streak_protected = False
            stats.current_streak += 1
            incremented = True
        else:
            stats.current_streak = 1
            incremented = True

    if stats.current_streak > stats.best_streak:
        stats.best_streak = stats.current_streak

    stats.last_activity_date = today
    return incremented, stats.current_streak

async def add_xp_and_currency(
    db: AsyncSession,
    user_id: int,
    xp_amount: int,
    currency_amount: int = 0
) -> Tuple[UserStats, bool, int]:
    """Awards XP and currency, recalculates level, triggers level-up bonuses."""
    res = await db.execute(select(UserStats).where(UserStats.user_id == user_id))
    stats = res.scalar_one_or_none()
    if not stats:
        stats = UserStats(user_id=user_id, xp=0, level=1, currency=50)
        db.add(stats)
        await db.flush()

    old_level = stats.level
    stats.xp += xp_amount
    stats.currency += currency_amount

    new_level, _, _ = get_level_info(stats.xp)
    level_up = new_level > old_level
    if level_up:
        stats.level = new_level
        # Level up currency bonus (+25 per level gained)
        stats.currency += (new_level - old_level) * 25

    return stats, level_up, new_level

async def update_quest_progress(
    db: AsyncSession,
    user_id: int,
    quest_type: str,
    increment: int = 1
) -> List[UserQuest]:
    """Updates daily quests of given type for today."""
    today = date.today()
    # Find all quests matching quest_type
    res = await db.execute(select(DailyQuest).where(DailyQuest.quest_type == quest_type))
    quests = res.scalars().all()
    
    updated_quests = []
    for q in quests:
        # Check user quest
        uq_res = await db.execute(
            select(UserQuest).where(
                UserQuest.user_id == user_id,
                UserQuest.quest_id == q.id,
                UserQuest.quest_date == today
            )
        )
        uq = uq_res.scalar_one_or_none()
        if not uq:
            uq = UserQuest(
                user_id=user_id,
                quest_id=q.id,
                quest_date=today,
                progress=0,
                completed=False,
                claimed=False
            )
            db.add(uq)
            await db.flush()
            
        if not uq.completed:
            uq.progress += increment
            if uq.progress >= q.target:
                uq.progress = q.target
                uq.completed = True
                # Auto-grant reward on completion
                await add_xp_and_currency(db, user_id, q.reward_xp, q.reward_currency)
            updated_quests.append(uq)
            
    return updated_quests

async def check_achievements(db: AsyncSession, user_id: int) -> Optional[AchievementOut]:
    """Checks conditions and unlocks newly achieved achievements."""
    res = await db.execute(select(UserStats).where(UserStats.user_id == user_id))
    stats = res.scalar_one_or_none()
    if not stats:
        return None

    # Get already unlocked achievement IDs
    unlocked_res = await db.execute(
        select(UserAchievement.achievement_id).where(UserAchievement.user_id == user_id)
    )
    unlocked_ids = set(unlocked_res.scalars().all())

    # Get all achievements
    all_ach_res = await db.execute(select(Achievement))
    all_achievements = all_ach_res.scalars().all()

    first_new_unlock = None

    for ach in all_achievements:
        if ach.id in unlocked_ids:
            continue
            
        condition_met = False
        if ach.condition_type == "tasks_count" and stats.total_tasks_completed >= ach.condition_value:
            condition_met = True
        elif ach.condition_type == "streak_days" and stats.current_streak >= ach.condition_value:
            condition_met = True
        elif ach.condition_type == "proposals_count" and stats.total_proposals_made >= ach.condition_value:
            condition_met = True
        elif ach.condition_type == "confirmed_count" and stats.total_translations_confirmed >= ach.condition_value:
            condition_met = True
        elif ach.condition_type == "xp_total" and stats.xp >= ach.condition_value:
            condition_met = True

        if condition_met:
            ua = UserAchievement(user_id=user_id, achievement_id=ach.id)
            db.add(ua)
            await add_xp_and_currency(db, user_id, ach.reward_xp, ach.reward_currency)
            if not first_new_unlock:
                first_new_unlock = AchievementOut(
                    id=ach.id,
                    code=ach.code,
                    name=ach.name,
                    description=ach.description,
                    icon=ach.icon,
                    reward_xp=ach.reward_xp,
                    reward_currency=ach.reward_currency,
                    unlocked=True,
                    unlocked_at=datetime.utcnow()
                )

    return first_new_unlock
