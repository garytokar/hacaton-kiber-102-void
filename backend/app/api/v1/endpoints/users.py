from datetime import date, timedelta
from typing import List, Dict, Any, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload
from app.core.database import get_db
from app.models.models import User, UserStats, UserAnswer, School, SchoolClass, DictionaryEntry
from app.schemas.schemas import UserOut, UserStatsOut, UserUpdate
from app.api.deps import get_current_user
from app.api.v1.endpoints.auth import format_user_out
from app.core.cosmetics import parse_user_cosmetics, serialize_user_cosmetics
from pydantic import BaseModel
import random

class ShopBuyRequest(BaseModel):
    item_id: str

class ShopEquipRequest(BaseModel):
    item_id: str
    category: str = "frame"  # frame or title

SHOP_CATALOG = [
    {
        "id": "amulet_streak",
        "name": "Амулет Салавата (Защита серии)",
        "description": "Священный оберег: замораживает серию дней при пропуске ежедневного задания.",
        "price": 50,
        "category": "utility",
        "icon": "shield",
    },
    {
        "id": "frame_gold",
        "name": "Солярная золотая оправа",
        "description": "Благородная золотая оправа с солярным башкирским знаком для вашего аватара.",
        "price": 75,
        "category": "frame",
        "icon": "sun",
    },
    {
        "id": "frame_sapphire",
        "name": "Сапфировое сияние Урала",
        "description": "Глубокое неоновое сапфировое свечение вокруг аватара в цветах платформы.",
        "price": 100,
        "category": "frame",
        "icon": "sparkles",
    },
    {
        "id": "frame_emerald",
        "name": "Изумрудный венец Агидели",
        "description": "Малахитово-изумрудная рамка Хранителя заповедной башкирской природы.",
        "price": 120,
        "category": "frame",
        "icon": "gem",
    },
    {
        "id": "title_aksakal",
        "name": "Почётный титул «Аксакал перевода»",
        "description": "Особый статус старейшины и золотой значок мудреца возле вашего имени.",
        "price": 150,
        "category": "title",
        "icon": "award",
    },
    {
        "id": "wisdom_scroll",
        "name": "Свиток народной мудрости",
        "description": "Открывает редкую старинную башкирскую пословицу в ваш персональный словарь.",
        "price": 30,
        "category": "lore",
        "icon": "scroll",
    },
]

WISDOM_PROVERBS = [
    {
        "word": "Ат аяғы ет етмәҫ, ир аяғы ер итмәҫ",
        "translation": "Конь ног не стопчет, джигит землю не потеряет (Смелость преодолевает все преграды)",
        "context": "Башкирский героический фольклор"
    },
    {
        "word": "Батыр бер үлер, ҡурҡаҡ мең үлер",
        "translation": "Герой умирает один раз, трус — тысячу раз",
        "context": "Народная мудрость о чести и отваге"
    },
    {
        "word": "Ағас япрағы менән, кеше дуҫтары менән",
        "translation": "Дерево сильно своей листвой, а человек — своими верными друзьями",
        "context": "Башкирская пословица о дружбе"
    },
    {
        "word": "Кешегә игелек итһәң, үҙеңә игелек ҡайтыр",
        "translation": "Сделаешь человеку добро — добро к тебе вернётся",
        "context": "Заветы старейшин"
    },
    {
        "word": "Тырышҡан таш ярыр, ялҡау ҡул ҡаушырыр",
        "translation": "Старательный и камень расколет, а ленивый лишь руки сложит",
        "context": "Пословица об усердном учении"
    }
]

router = APIRouter()

@router.get("/me", response_model=UserOut)
async def get_current_user_profile(user: User = Depends(get_current_user)):
    return format_user_out(user)

@router.patch("/me", response_model=UserOut)
async def update_current_user_profile(
    req: UserUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    if req.full_name is not None:
        user.full_name = req.full_name
    if req.avatar is not None:
        user.avatar = req.avatar
    if req.school_id is not None:
        user.school_id = req.school_id
    if req.class_id is not None:
        user.class_id = req.class_id

    await db.commit()

    stmt = select(User).options(
        selectinload(User.stats),
        selectinload(User.school),
        selectinload(User.school_class)
    ).where(User.id == user.id)
    res = await db.execute(stmt)
    refreshed_user = res.scalar_one()
    return format_user_out(refreshed_user)

@router.get("/me/stats", response_model=UserStatsOut)
async def get_current_user_stats(user: User = Depends(get_current_user)):
    user_out = format_user_out(user)
    if not user_out.stats:
        raise HTTPException(status_code=404, detail="Статистика не найдена")
    return user_out.stats

@router.get("/me/activity")
async def get_user_activity(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Returns daily activity counts for the last 60 days (GitHub Contributions style)."""
    end_date = date.today()
    start_date = end_date - timedelta(days=60)

    # Count answers per day
    res = await db.execute(
        select(
            func.date(UserAnswer.created_at).label("day"),
            func.count(UserAnswer.id).label("count")
        ).where(
            UserAnswer.user_id == user.id,
            UserAnswer.created_at >= start_date
        ).group_by(func.date(UserAnswer.created_at))
    )
    raw_counts = {str(row[0]): row[1] for row in res.all()}

    calendar = []
    current = start_date
    while current <= end_date:
        d_str = str(current)
        calendar.append({
            "date": d_str,
            "count": raw_counts.get(d_str, 0)
        })
        current += timedelta(days=1)

    return calendar

@router.post("/me/protect-streak")
async def buy_streak_protection(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Buy protective amulet (streak freeze) for 50 Ҡурай currency."""
    if not user.stats:
        raise HTTPException(status_code=404, detail="Статистика пользователя отсутствует")
    if user.stats.streak_protected:
        return {"success": False, "message": "Защита серии уже активна"}
    if user.stats.currency < 50:
        raise HTTPException(status_code=400, detail="Недостаточно валюты Ҡурай (нужно 50)")

    user.stats.currency -= 50
    user.stats.streak_protected = True
    await db.commit()

    return {"success": True, "message": "Оберег серии активирован! Ваша серия в безопасности.", "currency": user.stats.currency}

class LangPairUpdate(BaseModel):
    preferred_lang_pair: Optional[str] = None
    lang_pair: Optional[str] = None
    language_pair: Optional[str] = None
    lang: Optional[str] = None

@router.get("/me/lang-pair")
async def get_user_lang_pair(user: User = Depends(get_current_user)):
    lp = getattr(user, "preferred_lang_pair", None) or "ba-ru"
    return {
        "preferred_lang_pair": lp,
        "lang_pair": lp,
        "available_pairs": ["ba-ru", "ru-ba", "ba-kk", "ba-en"]
    }

@router.put("/me/lang-pair")
@router.patch("/me/lang-pair")
async def update_user_lang_pair(
    req: LangPairUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    val = req.preferred_lang_pair or req.lang_pair or req.language_pair or req.lang
    if not val:
        raise HTTPException(status_code=400, detail="Не указана языковая пара")
    val = val.lower().strip()
    if val not in ["ba-ru", "ru-ba", "ba-kk", "ba-en", "en-ba"]:
        raise HTTPException(status_code=400, detail="Поддерживаются пары ba-ru, ru-ba, ba-kk, ba-en")
    user.preferred_lang_pair = val
    await db.commit()
    return {
        "success": True,
        "preferred_lang_pair": val,
        "lang_pair": val
    }

@router.get("/me/shop/items")
async def get_shop_items(user: User = Depends(get_current_user)):
    """Return all shop items with user ownership and equip state."""
    cosmetics = parse_user_cosmetics(user.avatar)
    unlocked = set(cosmetics.get("unlocked", []))
    equipped_frame = cosmetics.get("equipped_frame")
    equipped_title = cosmetics.get("equipped_title")
    streak_protected = user.stats.streak_protected if user.stats else False

    result = []
    for item in SHOP_CATALOG:
        is_owned = False
        is_equipped = False

        if item["id"] == "amulet_streak":
            is_owned = streak_protected
            is_equipped = False
        elif item["id"] == "wisdom_scroll":
            is_owned = False
            is_equipped = False
        else:
            is_owned = item["id"] in unlocked
            if item["category"] == "frame":
                is_equipped = (equipped_frame == item["id"])
            elif item["category"] == "title":
                is_equipped = (equipped_title == item["id"])

        result.append({
            **item,
            "owned": is_owned,
            "equipped": is_equipped
        })

    return {
        "items": result,
        "currency": user.stats.currency if user.stats else 0,
        "cosmetics": cosmetics
    }

@router.post("/me/shop/buy")
async def buy_shop_item(
    req: ShopBuyRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Purchase an item from the Guardian Shop using Ҡурай coins."""
    if not user.stats:
        raise HTTPException(status_code=404, detail="Статистика пользователя не найдена")

    item = next((i for i in SHOP_CATALOG if i["id"] == req.item_id), None)
    if not item:
        raise HTTPException(status_code=404, detail="Товар не найден в лавке")

    cosmetics = parse_user_cosmetics(user.avatar)

    # Check already owned for permanent items
    if item["id"] == "amulet_streak":
        if user.stats.streak_protected:
            return {"success": False, "message": "Защита серии уже активна на вашем аккаунте"}
    elif item["id"] != "wisdom_scroll":
        if item["id"] in cosmetics.get("unlocked", []):
            return {"success": False, "message": "Этот предмет уже у вас в инвентаре"}

    if user.stats.currency < item["price"]:
        raise HTTPException(
            status_code=400,
            detail=f"Недостаточно монет Ҡурай: требуется {item['price']}, а у вас {user.stats.currency}"
        )

    # Deduct currency
    user.stats.currency -= item["price"]

    # Apply effect
    if item["id"] == "amulet_streak":
        user.stats.streak_protected = True
        msg = "Амулет Салавата активирован! Ваша серия дней защищена."
    elif item["id"] == "wisdom_scroll":
        proverb = random.choice(WISDOM_PROVERBS)
        new_word = DictionaryEntry(
            user_id=user.id,
            word=proverb["word"],
            translation=proverb["translation"],
            context=proverb["context"],
            difficulty="hard",
            is_favorite=True
        )
        db.add(new_word)
        msg = f"Мудрость предков получена: «{proverb['word']}» добавлено в ваш словарь!"
    else:
        # Cosmetic frame or title
        if item["id"] not in cosmetics["unlocked"]:
            cosmetics["unlocked"].append(item["id"])
        if item["category"] == "frame":
            cosmetics["equipped_frame"] = item["id"]
        elif item["category"] == "title":
            cosmetics["equipped_title"] = item["id"]
        user.avatar = serialize_user_cosmetics(cosmetics)
        msg = f"Поздравляем! {item['name']} успешно приобретён и экипирован!"

    await db.commit()

    return {
        "success": True,
        "message": msg,
        "currency": user.stats.currency,
        "cosmetics": cosmetics,
        "user": format_user_out(user)
    }

@router.post("/me/shop/equip")
async def equip_shop_item(
    req: ShopEquipRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Equip or unequip an unlocked cosmetic item."""
    cosmetics = parse_user_cosmetics(user.avatar)

    if req.item_id and req.item_id not in cosmetics.get("unlocked", []):
        raise HTTPException(status_code=400, detail="Этот предмет ещё не открыт вами")

    if req.category == "frame":
        # Toggle or set
        if cosmetics.get("equipped_frame") == req.item_id:
            cosmetics["equipped_frame"] = None
        else:
            cosmetics["equipped_frame"] = req.item_id
    elif req.category == "title":
        if cosmetics.get("equipped_title") == req.item_id:
            cosmetics["equipped_title"] = None
        else:
            cosmetics["equipped_title"] = req.item_id

    user.avatar = serialize_user_cosmetics(cosmetics)
    await db.commit()

    return {
        "success": True,
        "cosmetics": cosmetics,
        "user": format_user_out(user)
    }

@router.get("/schools")
async def list_schools(db: AsyncSession = Depends(get_db)):
    """Returns schools and their classes for selection."""
    stmt = select(School).options(selectinload(School.classes))
    res = await db.execute(stmt)
    schools = res.scalars().all()
    return [
        {
            "id": s.id,
            "name": s.name,
            "city": s.city,
            "classes": [{"id": c.id, "name": c.name} for c in s.classes]
        }
        for s in schools
    ]

@router.get("/me/team")
async def get_my_team_info(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Return user's school team members, team XP, collective goal, and rank."""
    if not user.school_id:
        return {
            "has_school": False,
            "message": "Вы пока не прикреплены к школе. Выберите школу в настройках или профиле."
        }

    school = (await db.execute(select(School).where(School.id == user.school_id))).scalar_one_or_none()
    if not school:
        return {"has_school": False}

    # Fetch school classmates and teammates
    teammates_res = await db.execute(
        select(User)
        .options(selectinload(User.stats))
        .where(User.school_id == user.school_id)
        .order_by(User.id)
    )
    teammates = teammates_res.scalars().all()

    total_team_xp = sum(t.stats.xp for t in teammates if t.stats)
    total_confirmed = sum(t.stats.total_translations_confirmed for t in teammates if t.stats)

    members_out = [
        {
            "id": t.id,
            "username": t.username,
            "full_name": t.full_name or t.username,
            "avatar": t.avatar,
            "role": t.role,
            "xp": t.stats.xp if t.stats else 0,
            "level": t.stats.level if t.stats else 1,
            "is_me": (t.id == user.id),
            "cosmetics": parse_user_cosmetics(t.avatar)
        }
        for t in teammates
    ]

    return {
        "has_school": True,
        "school": {
            "id": school.id,
            "name": school.name,
            "city": school.city,
            "rank": 1,
            "total_xp": total_team_xp,
            "members_count": len(members_out),
            "class_name": user.school_class.name if user.school_class else None
        },
        "collective_goal": {
            "title": "Коллективная миссия школы: 100 подтверждённых переводов за неделю",
            "target": 100,
            "current": min(total_confirmed + 42, 100),
            "percentage": min(round(((total_confirmed + 42) / 100) * 100), 100),
            "reward_xp": 150,
            "reward_currency": 50
        },
        "members": members_out
    }
