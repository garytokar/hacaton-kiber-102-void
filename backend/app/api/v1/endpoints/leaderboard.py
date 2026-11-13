from typing import List, Optional
from fastapi import APIRouter, Depends, Query
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, desc, func
from sqlalchemy.orm import selectinload
from app.core.database import get_db
from app.models.models import User, UserStats, School, Team, TeamMember
from app.schemas.schemas import LeaderboardUser, TeamLeaderboardEntry, SchoolLeaderboardEntry
from app.services.gamification import get_level_info

router = APIRouter()

@router.get("", response_model=List[LeaderboardUser])
async def get_user_leaderboard(
    period: str = Query("all_time", pattern="^(weekly|week|monthly|month|all_time|all)$"),
    type: str = Query("xp", pattern="^(xp|quality)$"),
    limit: int = Query(50, le=100),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(User).join(User.stats).options(
        selectinload(User.stats),
        selectinload(User.school)
    )

    if type == "quality":
        # Order by total_translations_confirmed + total_accepted_proposals
        stmt = stmt.order_by(
            desc(UserStats.total_translations_confirmed + UserStats.total_accepted_proposals),
            desc(UserStats.xp)
        )
    else:
        stmt = stmt.order_by(desc(UserStats.xp))

    stmt = stmt.limit(limit)
    res = await db.execute(stmt)
    users = res.scalars().all()

    out = []
    for rank, u in enumerate(users, start=1):
        if not u.stats:
            continue
        lvl, title, _ = get_level_info(u.stats.xp)
        out.append(
            LeaderboardUser(
                rank=rank,
                user_id=u.id,
                username=u.username,
                full_name=u.full_name,
                avatar=u.avatar,
                level=lvl,
                level_title=title,
                xp=u.stats.xp,
                streak=u.stats.current_streak,
                confirmed_count=u.stats.total_translations_confirmed,
                school_name=u.school.name if u.school else "Индивидуальный участник"
            )
        )
    return out

@router.get("/teams", response_model=List[TeamLeaderboardEntry])
async def get_team_leaderboard(db: AsyncSession = Depends(get_db)):
    teams_res = await db.execute(select(Team))
    teams = teams_res.scalars().all()

    team_stats = []
    for t in teams:
        members_res = await db.execute(
            select(UserStats).join(User, User.id == UserStats.user_id).join(
                TeamMember, TeamMember.user_id == User.id
            ).where(TeamMember.team_id == t.id)
        )
        stats_list = members_res.scalars().all()
        m_count = len(stats_list)
        total_xp = sum(s.xp for s in stats_list)
        total_confirmed = sum(s.total_translations_confirmed for s in stats_list)

        team_stats.append({
            "team_id": t.id,
            "name": t.name,
            "emblem": t.emblem,
            "members_count": m_count,
            "total_xp": total_xp,
            "total_confirmed": total_confirmed
        })

    team_stats.sort(key=lambda x: x["total_xp"], reverse=True)

    return [
        TeamLeaderboardEntry(
            rank=idx + 1,
            team_id=ts["team_id"],
            name=ts["name"],
            emblem=ts["emblem"],
            members_count=ts["members_count"],
            total_xp=ts["total_xp"],
            total_confirmed=ts["total_confirmed"]
        )
        for idx, ts in enumerate(team_stats)
    ]

@router.get("/schools", response_model=List[SchoolLeaderboardEntry])
async def get_school_leaderboard(db: AsyncSession = Depends(get_db)):
    schools_res = await db.execute(select(School))
    schools = schools_res.scalars().all()

    school_stats = []
    for s in schools:
        stats_res = await db.execute(
            select(UserStats).join(User, User.id == UserStats.user_id).where(User.school_id == s.id)
        )
        stats_list = stats_res.scalars().all()
        st_count = len(stats_list)
        total_xp = sum(st.xp for st in stats_list)
        total_conf = sum(st.total_translations_confirmed for st in stats_list)

        school_stats.append({
            "school_id": s.id,
            "name": s.name,
            "city": s.city,
            "students_count": st_count,
            "total_xp": total_xp,
            "total_confirmed": total_conf
        })

    school_stats.sort(key=lambda x: x["total_xp"], reverse=True)

    return [
        SchoolLeaderboardEntry(
            rank=idx + 1,
            school_id=ss["school_id"],
            name=ss["name"],
            city=ss["city"],
            students_count=ss["students_count"],
            total_xp=ss["total_xp"],
            total_confirmed=ss["total_confirmed"]
        )
        for idx, ss in enumerate(school_stats)
    ]

@router.get("/classes")
async def get_classes_leaderboard(
    limit: int = Query(20, le=100),
    db: AsyncSession = Depends(get_db)
):
    """Ranks school classes by total XP of their students (Раздел 4.5 ТЗ)."""
    from app.models.models import SchoolClass
    stmt = (
        select(
            SchoolClass.id.label("class_id"),
            SchoolClass.name.label("class_name"),
            School.id.label("school_id"),
            School.name.label("school_name"),
            func.count(User.id).label("students_count"),
            func.coalesce(func.sum(UserStats.xp), 0).label("total_xp"),
            func.coalesce(func.sum(UserStats.total_translations_confirmed), 0).label("total_confirmed")
        )
        .join(School, School.id == SchoolClass.school_id)
        .outerjoin(User, User.class_id == SchoolClass.id)
        .outerjoin(UserStats, UserStats.user_id == User.id)
        .group_by(SchoolClass.id, SchoolClass.name, School.id, School.name)
        .order_by(desc(func.coalesce(func.sum(UserStats.xp), 0)))
        .limit(limit)
    )
    res = await db.execute(stmt)
    rows = res.all()

    out = []
    for rank, r in enumerate(rows, start=1):
        out.append({
            "rank": rank,
            "class_id": r.class_id,
            "class_name": r.class_name,
            "name": f"{r.class_name} ({r.school_name})",
            "school_id": r.school_id,
            "school_name": r.school_name,
            "students_count": r.students_count,
            "total_xp": int(r.total_xp),
            "total_confirmed": int(r.total_confirmed)
        })
    return out

@router.get("/vocabulary")
async def get_vocabulary_leaderboard(
    limit: int = Query(50, le=100),
    db: AsyncSession = Depends(get_db)
):
    """Ranks users by vocabulary size and confirmed contributions (Раздел 4.5 ТЗ)."""
    from app.models.models import DictionaryEntry
    stmt = (
        select(
            User.id.label("user_id"),
            User.username.label("username"),
            User.full_name.label("full_name"),
            School.name.label("school_name"),
            UserStats.xp.label("xp"),
            func.count(DictionaryEntry.id).label("vocab_count"),
            UserStats.total_translations_confirmed.label("confirmed_count")
        )
        .join(UserStats, UserStats.user_id == User.id)
        .outerjoin(School, School.id == User.school_id)
        .outerjoin(DictionaryEntry, DictionaryEntry.user_id == User.id)
        .group_by(User.id, User.username, User.full_name, School.name, UserStats.xp, UserStats.total_translations_confirmed)
        .order_by(desc(func.count(DictionaryEntry.id) + UserStats.total_translations_confirmed), desc(UserStats.xp))
        .limit(limit)
    )
    res = await db.execute(stmt)
    rows = res.all()

    out = []
    for rank, r in enumerate(rows, start=1):
        lvl, title, _ = get_level_info(r.xp or 0)
        out.append({
            "rank": rank,
            "user_id": r.user_id,
            "username": r.username,
            "full_name": r.full_name or r.username,
            "school_name": r.school_name or "Республика Башкортостан",
            "vocabulary_count": r.vocab_count,
            "total_confirmed": r.confirmed_count,
            "xp": r.xp,
            "level": lvl,
            "level_title": title
        })
    return out
