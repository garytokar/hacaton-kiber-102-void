import pytest
from app.services.gamification import get_level_info, update_user_streak, LEVEL_THRESHOLDS
from app.models.models import UserStats
from datetime import date, timedelta

def test_all_10_level_thresholds():
    assert len(LEVEL_THRESHOLDS) == 10

    # 1. Начинающий (0 XP)
    lvl, title, _ = get_level_info(0)
    assert lvl == 1 and title == "Начинающий"

    # 2. Ученик (100 XP)
    lvl, title, _ = get_level_info(100)
    assert lvl == 2 and title == "Ученик"

    # 3. Искатель слов (250 XP)
    lvl, title, _ = get_level_info(250)
    assert lvl == 3 and title == "Искатель слов"

    # 4. Знаток (500 XP)
    lvl, title, _ = get_level_info(500)
    assert lvl == 4 and title == "Знаток"

    # 5. Переводчик (900 XP)
    lvl, title, _ = get_level_info(900)
    assert lvl == 5 and title == "Переводчик"

    # 6. Хранитель (1500 XP)
    lvl, title, _ = get_level_info(1500)
    assert lvl == 6 and title == "Хранитель"

    # 7. Мастер языка (2500 XP)
    lvl, title, _ = get_level_info(2500)
    assert lvl == 7 and title == "Мастер языка"

    # 8. Наставник (4000 XP)
    lvl, title, _ = get_level_info(4000)
    assert lvl == 8 and title == "Наставник"

    # 9. Мудрец (6500 XP)
    lvl, title, _ = get_level_info(6500)
    assert lvl == 9 and title == "Мудрец"

    # 10. Легенда языка (10000 XP)
    lvl, title, _ = get_level_info(12000)
    assert lvl == 10 and title == "Легенда языка"

@pytest.mark.asyncio
async def test_streak_increment_and_reset():
    today = date.today()
    stats = UserStats(user_id=999, current_streak=3, best_streak=5, last_activity_date=today - timedelta(days=1))

    # Activity today -> increments streak
    incr, cur = await update_user_streak(stats, today)
    assert incr is True
    assert cur == 4
    assert stats.current_streak == 4

    # Activity same day again -> does not increment second time
    incr2, cur2 = await update_user_streak(stats, today)
    assert incr2 is False
    assert cur2 == 4

    # Missed days with streak protection
    stats.last_activity_date = today - timedelta(days=3)
    stats.streak_protected = True
    incr3, cur3 = await update_user_streak(stats, today)
    assert incr3 is True
    assert cur3 == 5  # Saved and incremented
    assert stats.streak_protected is False  # Amulet consumed

    # Missed days without protection
    stats.last_activity_date = today - timedelta(days=3)
    stats.streak_protected = False
    incr4, cur4 = await update_user_streak(stats, today)
    assert cur4 == 1  # Reset to 1

@pytest.mark.asyncio
async def test_streak_protection_purchase_endpoint(client, student_token):
    res = await client.post(
        "/api/users/me/protect-streak",
        headers={"Authorization": f"Bearer {student_token}"}
    )
    # Either successfully bought or already active
    assert res.status_code in [200, 400]
