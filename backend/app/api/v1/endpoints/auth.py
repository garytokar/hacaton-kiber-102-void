from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from app.core.database import get_db
from app.core.security import verify_password, get_password_hash, create_access_token
from app.core.cosmetics import parse_user_cosmetics
from app.models.models import User, UserStats, School, SchoolClass
from app.schemas.schemas import UserRegister, UserLogin, Token, UserOut, UserStatsOut
from app.services.gamification import get_level_info
from app.api.deps import get_current_user

router = APIRouter()

def format_user_out(user: User) -> UserOut:
    stats_out = None
    if user.stats:
        lvl, title, next_xp = get_level_info(user.stats.xp)
        stats_out = UserStatsOut(
            user_id=user.stats.user_id,
            xp=user.stats.xp,
            level=lvl,
            level_title=title,
            next_level_xp=next_xp,
            currency=user.stats.currency,
            current_streak=user.stats.current_streak,
            best_streak=user.stats.best_streak,
            streak_protected=user.stats.streak_protected,
            total_tasks_completed=user.stats.total_tasks_completed,
            total_translations_confirmed=user.stats.total_translations_confirmed,
            total_proposals_made=user.stats.total_proposals_made,
            total_accepted_proposals=user.stats.total_accepted_proposals
        )

    return UserOut(
        id=user.id,
        email=user.email,
        username=user.username,
        full_name=user.full_name,
        avatar=user.avatar,
        role=user.role,
        school_id=user.school_id,
        school_name=user.school.name if user.school else None,
        class_id=user.class_id,
        class_name=user.school_class.name if user.school_class else None,
        created_at=user.created_at,
        stats=stats_out,
        cosmetics=parse_user_cosmetics(user.avatar)
    )

@router.post("/register", response_model=Token)
async def register(req: UserRegister, db: AsyncSession = Depends(get_db)):
    # Check email exists
    res = await db.execute(select(User).where(User.email == req.email.lower()))
    if res.scalar_one_or_none():
        raise HTTPException(status_code=400, detail="Пользователь с таким email уже существует")
    
    # Check username exists
    res = await db.execute(select(User).where(User.username == req.username))
    if res.scalar_one_or_none():
        raise HTTPException(status_code=400, detail="Пользователь с таким именем уже существует")

    new_user = User(
        email=req.email.lower(),
        username=req.username,
        password_hash=get_password_hash(req.password),
        full_name=req.full_name or req.username,
        avatar="avatar_default",
        role=req.role if req.role in ["student", "teacher", "expert", "admin"] else "student",
        school_id=req.school_id,
        class_id=req.class_id
    )
    db.add(new_user)
    await db.flush()

    # Initial stats
    new_stats = UserStats(
        user_id=new_user.id,
        xp=0,
        level=1,
        currency=50,
        current_streak=0,
        best_streak=0,
    )
    db.add(new_stats)
    await db.commit()

    # Reload with relationships
    stmt = select(User).options(
        selectinload(User.stats),
        selectinload(User.school),
        selectinload(User.school_class)
    ).where(User.id == new_user.id)
    user_res = await db.execute(stmt)
    user = user_res.scalar_one()

    token = create_access_token(user.id)
    return Token(access_token=token, token_type="bearer", user=format_user_out(user))

@router.post("/login", response_model=Token)
async def login(req: UserLogin, db: AsyncSession = Depends(get_db)):
    ident = req.email.lower().strip()
    stmt = select(User).options(
        selectinload(User.stats),
        selectinload(User.school),
        selectinload(User.school_class)
    ).where((User.email == ident) | (User.username == ident))
    res = await db.execute(stmt)
    user = res.scalar_one_or_none()

    if not user:
        # Support demo aliases student1, teacher1, expert1, admin
        alias_map = {
            "student1": "student_timur@yanlander.ru",
            "teacher1": "teacher_aygul@yanlander.ru",
            "expert1": "expert@yanlander.ru",
            "admin": "admin@yanlander.ru",
        }
        if ident in alias_map:
            res = await db.execute(select(User).options(
                selectinload(User.stats),
                selectinload(User.school),
                selectinload(User.school_class)
            ).where(User.email == alias_map[ident]))
            user = res.scalar_one_or_none()

    if not user or not verify_password(req.password, user.password_hash):
        raise HTTPException(status_code=400, detail="Неверный email или пароль")

    token = create_access_token(user.id)
    return Token(access_token=token, token_type="bearer", user=format_user_out(user))

@router.get("/me", response_model=UserOut)
async def get_me(user: User = Depends(get_current_user)):
    return format_user_out(user)
