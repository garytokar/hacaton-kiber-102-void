import pytest
import uuid
from httpx import AsyncClient
from app.services.gamification import get_level_info, update_user_streak, LEVEL_THRESHOLDS

# ============================================================================
# 1. AUTHENTICATION & PROFILE TESTS (10 tests)
# ============================================================================

@pytest.mark.asyncio
async def test_auth_me_student_profile(client: AsyncClient, student_token: str):
    res = await client.get("/api/auth/me", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    data = res.json()
    assert data["role"] == "student"
    assert "email" in data
    assert "stats" in data

@pytest.mark.asyncio
async def test_auth_me_admin_profile(client: AsyncClient, admin_token: str):
    res = await client.get("/api/auth/me", headers={"Authorization": f"Bearer {admin_token}"})
    assert res.status_code == 200
    data = res.json()
    assert data["role"] == "admin"

@pytest.mark.asyncio
async def test_auth_me_unauthorized(client: AsyncClient):
    res = await client.get("/api/auth/me")
    assert res.status_code in (401, 403)

@pytest.mark.asyncio
async def test_auth_me_invalid_token(client: AsyncClient):
    res = await client.get("/api/auth/me", headers={"Authorization": "Bearer fake_token_xyz"})
    assert res.status_code in (401, 403)

@pytest.mark.asyncio
async def test_auth_register_new_student(client: AsyncClient):
    uname = f"suite_stu_{uuid.uuid4().hex[:6]}"
    res = await client.post("/api/auth/register", json={
        "email": f"{uname}@yanlander.ba",
        "username": uname,
        "password": "Password123!",
        "full_name": "Хранитель Тестовый",
        "role": "student"
    })
    assert res.status_code == 200
    data = res.json()
    assert "access_token" in data
    assert data["user"]["username"] == uname

@pytest.mark.asyncio
async def test_auth_register_duplicate_email(client: AsyncClient):
    uname = f"dup_{uuid.uuid4().hex[:6]}"
    email = f"{uname}@yanlander.ru"
    await client.post("/api/auth/register", json={
        "email": email,
        "username": uname,
        "password": "Password123!"
    })
    res_dup = await client.post("/api/auth/register", json={
        "email": email,
        "username": f"diff_{uname}",
        "password": "Password123!"
    })
    assert res_dup.status_code == 400

@pytest.mark.asyncio
async def test_auth_login_with_email(client: AsyncClient):
    res = await client.post("/api/auth/login", json={
        "email": "student_timur@yanlander.ru",
        "password": "password123"
    })
    assert res.status_code == 200
    data = res.json()
    assert "access_token" in data
    assert data["user"]["email"] == "student_timur@yanlander.ru"

@pytest.mark.asyncio
async def test_auth_login_bad_password(client: AsyncClient):
    res = await client.post("/api/auth/login", json={
        "email": "student_timur@yanlander.ru",
        "password": "wrong_password_123"
    })
    assert res.status_code == 400

@pytest.mark.asyncio
async def test_auth_login_nonexistent_email(client: AsyncClient):
    res = await client.post("/api/auth/login", json={
        "email": "nobody_exists_here_never@yanlander.ru",
        "password": "password"
    })
    assert res.status_code == 400

@pytest.mark.asyncio
async def test_auth_token_contains_user_id(client: AsyncClient, student_token: str):
    res = await client.get("/api/auth/me", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    assert "id" in res.json()


# ============================================================================
# 2. DAILY FRAGMENTS & ARCHIVE (10 tests)
# ============================================================================

@pytest.mark.asyncio
async def test_get_daily_fragment_structure(client: AsyncClient):
    res = await client.get("/api/fragments/daily")
    assert res.status_code == 200
    data = res.json()
    assert "id" in data
    assert "text_ba" in data
    assert "translations" in data
    assert len(data["translations"]) > 0

@pytest.mark.asyncio
async def test_list_fragments_pagination(client: AsyncClient):
    res = await client.get("/api/fragments?limit=3&offset=0")
    assert res.status_code == 200
    data = res.json()
    assert isinstance(data, list)
    assert len(data) <= 3

@pytest.mark.asyncio
async def test_list_fragments_offset(client: AsyncClient):
    res1 = await client.get("/api/fragments?limit=2&offset=0")
    res2 = await client.get("/api/fragments?limit=2&offset=2")
    d1 = res1.json()
    d2 = res2.json()
    if d1 and d2:
        assert d1[0]["id"] != d2[0]["id"]

@pytest.mark.asyncio
async def test_get_fragment_by_id_success(client: AsyncClient):
    res = await client.get("/api/fragments/1")
    assert res.status_code == 200
    data = res.json()
    assert data["id"] == 1
    assert "text_ba" in data
    assert "source_title" in data

@pytest.mark.asyncio
async def test_get_fragment_by_id_404(client: AsyncClient):
    res = await client.get("/api/fragments/9999999")
    assert res.status_code == 404

@pytest.mark.asyncio
async def test_fragment_contains_dialect_metadata(client: AsyncClient):
    res = await client.get("/api/fragments/1")
    data = res.json()
    assert "dialect" in data or "difficulty" in data

@pytest.mark.asyncio
async def test_fragment_translations_have_votes(client: AsyncClient):
    res = await client.get("/api/fragments/1")
    data = res.json()
    for t in data.get("translations", []):
        assert "votes_count" in t
        assert "reject_count" in t

@pytest.mark.asyncio
async def test_filter_fragments_by_category(client: AsyncClient):
    res = await client.get("/api/fragments?category=proverbs")
    assert res.status_code == 200
    assert isinstance(res.json(), list)

@pytest.mark.asyncio
async def test_filter_fragments_by_difficulty(client: AsyncClient):
    res = await client.get("/api/fragments?difficulty=medium")
    assert res.status_code == 200
    assert isinstance(res.json(), list)

@pytest.mark.asyncio
async def test_fragments_list_empty_filter_handled(client: AsyncClient):
    res = await client.get("/api/fragments?category=non_existent_category_1234")
    assert res.status_code == 200
    assert res.json() == []


# ============================================================================
# 3. TRANSLATIONS, VOTING & CONSENSUS (10 tests)
# ============================================================================

@pytest.mark.asyncio
async def test_propose_translation_success(client: AsyncClient, student_token: str):
    res = await client.post("/api/translations", json={
        "fragment_id": 2,
        "translation_text": "Новый вариант перевода для комплексного теста."
    }, headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    data = res.json()
    assert data["success"] is True
    assert "xp_awarded" in data
    assert "currency_awarded" in data

@pytest.mark.asyncio
async def test_propose_translation_empty_text_fails(client: AsyncClient, student_token: str):
    res = await client.post("/api/translations", json={
        "fragment_id": 2,
        "translation_text": ""
    }, headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code in (400, 422)

@pytest.mark.asyncio
async def test_propose_translation_invalid_fragment_fails(client: AsyncClient, student_token: str):
    res = await client.post("/api/translations", json={
        "fragment_id": 9999999,
        "translation_text": "Тестовый перевод"
    }, headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code in (400, 404, 422)

@pytest.mark.asyncio
async def test_vote_translation_confirm_flow(client: AsyncClient):
    uname = f"voter_{uuid.uuid4().hex[:6]}"
    reg = await client.post("/api/auth/register", json={
        "email": f"{uname}@test.ba",
        "username": uname,
        "password": "Password123!"
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    res = await client.post("/api/translations/1/vote", json={"vote_type": "confirm"}, headers=headers)
    assert res.status_code == 200
    assert res.json()["success"] is True

@pytest.mark.asyncio
async def test_vote_translation_reject_flow(client: AsyncClient):
    uname = f"voter_{uuid.uuid4().hex[:6]}"
    reg = await client.post("/api/auth/register", json={
        "email": f"{uname}@test.ba",
        "username": uname,
        "password": "Password123!"
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    res = await client.post("/api/translations/2/vote", json={"vote_type": "reject"}, headers=headers)
    assert res.status_code == 200
    assert res.json()["success"] is True

@pytest.mark.asyncio
async def test_duplicate_vote_rejected(client: AsyncClient):
    uname = f"voter_{uuid.uuid4().hex[:6]}"
    reg = await client.post("/api/auth/register", json={
        "email": f"{uname}@test.ba",
        "username": uname,
        "password": "Password123!"
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    await client.post("/api/translations/3/vote", json={"vote_type": "confirm"}, headers=headers)
    res_dup = await client.post("/api/translations/3/vote", json={"vote_type": "confirm"}, headers=headers)
    assert res_dup.status_code == 400
    assert "уже проголосовали" in res_dup.json()["detail"]

@pytest.mark.asyncio
async def test_vote_unauthorized_fails(client: AsyncClient):
    res = await client.post("/api/translations/1/vote", json={"vote_type": "confirm"})
    assert res.status_code in (401, 403)

@pytest.mark.asyncio
async def test_vote_invalid_vote_type_fails(client: AsyncClient, student_token: str):
    res = await client.post("/api/translations/1/vote", json={"vote_type": "invalid_vote"}, headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code in (400, 422)

@pytest.mark.asyncio
async def test_vote_nonexistent_translation_fails(client: AsyncClient, student_token: str):
    res = await client.post("/api/translations/9999999/vote", json={"vote_type": "confirm"}, headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code in (400, 404)

@pytest.mark.asyncio
async def test_get_translation_history_single(client: AsyncClient):
    res = await client.get("/api/translations/1/history")
    assert res.status_code == 200
    assert isinstance(res.json(), list)


# ============================================================================
# 4. TREASURY & DICTIONARY (10 tests)
# ============================================================================

@pytest.mark.asyncio
async def test_treasury_items_list(client: AsyncClient):
    res = await client.get("/api/stats/treasury")
    assert res.status_code == 200
    data = res.json()
    assert "items" in data
    assert "categories" in data
    assert len(data["items"]) > 0

@pytest.mark.asyncio
async def test_treasury_filter_category(client: AsyncClient):
    res = await client.get("/api/stats/treasury?category=proverbs")
    assert res.status_code == 200
    data = res.json()
    assert "items" in data

@pytest.mark.asyncio
async def test_dictionary_add_word(client: AsyncClient, student_token: str):
    uname = f"dict_{uuid.uuid4().hex[:6]}"
    res = await client.post("/api/dictionary", json={
        "word": f"ТестҺүҙ_{uname}",
        "translation": "Тестовое слово",
        "context": "Пример употребления"
    }, headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    assert res.json()["word"].startswith("ТестҺүҙ")

@pytest.mark.asyncio
async def test_dictionary_get_all(client: AsyncClient, student_token: str):
    res = await client.get("/api/dictionary", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    assert isinstance(res.json(), list)

@pytest.mark.asyncio
async def test_dictionary_toggle_favorite(client: AsyncClient, student_token: str):
    add_res = await client.post("/api/dictionary", json={
        "word": "Ҡояш",
        "translation": "Солнце",
        "context": "Яҡты ҡояш"
    }, headers={"Authorization": f"Bearer {student_token}"})
    w_id = add_res.json()["id"]

    fav_res = await client.patch(f"/api/dictionary/{w_id}/favorite", headers={"Authorization": f"Bearer {student_token}"})
    assert fav_res.status_code == 200
    assert fav_res.json()["is_favorite"] is True

@pytest.mark.asyncio
async def test_dictionary_delete_word(client: AsyncClient, student_token: str):
    add_res = await client.post("/api/dictionary", json={
        "word": "Ваҡытлы",
        "translation": "Временный",
        "context": "Удаляемое слово"
    }, headers={"Authorization": f"Bearer {student_token}"})
    w_id = add_res.json()["id"]

    del_res = await client.delete(f"/api/dictionary/{w_id}", headers={"Authorization": f"Bearer {student_token}"})
    assert del_res.status_code == 200

@pytest.mark.asyncio
async def test_dictionary_delete_nonexistent_word(client: AsyncClient, student_token: str):
    res = await client.delete("/api/dictionary/9999999", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code in (404, 400)

@pytest.mark.asyncio
async def test_dictionary_unauthorized_access(client: AsyncClient):
    res = await client.get("/api/dictionary")
    assert res.status_code in (401, 403)

@pytest.mark.asyncio
async def test_treasury_contains_audio_metadata(client: AsyncClient):
    res = await client.get("/api/stats/treasury")
    items = res.json().get("items", [])
    assert any("audio_url" in item or "text_ba" in item for item in items)

@pytest.mark.asyncio
async def test_treasury_pagination_limit(client: AsyncClient):
    res = await client.get("/api/stats/treasury?limit=5")
    assert res.status_code == 200
    items = res.json().get("items", [])
    assert len(items) <= 5


# ============================================================================
# 5. LEADERBOARD, QUESTS & EXPEDITIONS (10 tests)
# ============================================================================

@pytest.mark.asyncio
async def test_leaderboard_all_users(client: AsyncClient):
    res = await client.get("/api/leaderboard")
    assert res.status_code == 200
    data = res.json()
    assert isinstance(data, list)
    assert len(data) > 0

@pytest.mark.asyncio
async def test_leaderboard_league_filtering(client: AsyncClient):
    res = await client.get("/api/leaderboard?league=bronze")
    assert res.status_code == 200
    assert isinstance(res.json(), list)

@pytest.mark.asyncio
async def test_leaderboard_top_users_order(client: AsyncClient):
    res = await client.get("/api/leaderboard?limit=10")
    data = res.json()
    xps = [u.get("xp", 0) for u in data]
    assert xps == sorted(xps, reverse=True)

@pytest.mark.asyncio
async def test_get_daily_quests_list(client: AsyncClient, student_token: str):
    res = await client.get("/api/quests/daily", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    data = res.json()
    assert isinstance(data, list)
    if data:
        assert "name" in data[0]
        assert "reward_xp" in data[0]

@pytest.mark.asyncio
async def test_get_weekly_expedition_data(client: AsyncClient):
    res = await client.get("/api/quests/expedition")
    assert res.status_code == 200
    data = res.json()
    assert "target_count" in data
    assert "current_count" in data
    assert "title" in data

@pytest.mark.asyncio
async def test_expedition_has_progress_pct(client: AsyncClient):
    res = await client.get("/api/quests/expedition")
    data = res.json()
    assert "progress_percentage" in data

@pytest.mark.asyncio
async def test_achievements_catalog(client: AsyncClient):
    res = await client.get("/api/achievements")
    assert res.status_code == 200
    data = res.json()
    assert isinstance(data, list)
    assert len(data) >= 5

@pytest.mark.asyncio
async def test_achievements_with_user_token(client: AsyncClient, student_token: str):
    res = await client.get("/api/achievements", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    data = res.json()
    assert isinstance(data, list)
    assert "unlocked" in data[0]

@pytest.mark.asyncio
async def test_user_stats_level_info(client: AsyncClient, student_token: str):
    res = await client.get("/api/auth/me", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    data = res.json()
    assert "stats" in data
    assert "level" in data["stats"]
    assert "xp" in data["stats"]

@pytest.mark.asyncio
async def test_user_currency_balance(client: AsyncClient, student_token: str):
    res = await client.get("/api/auth/me", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 200
    data = res.json()
    assert "currency" in data["stats"]
    assert data["stats"]["currency"] >= 0


# ============================================================================
# 6. ADMIN, MODERATION & DATA EXPORTS (10 tests)
# ============================================================================

@pytest.mark.asyncio
async def test_admin_review_queue_expert(client: AsyncClient, expert_token: str):
    res = await client.get("/api/admin/review", headers={"Authorization": f"Bearer {expert_token}"})
    assert res.status_code == 200
    assert isinstance(res.json(), list)

@pytest.mark.asyncio
async def test_admin_review_queue_student_forbidden(client: AsyncClient, student_token: str):
    res = await client.get("/api/admin/review", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 403

@pytest.mark.asyncio
async def test_admin_users_list_admin(client: AsyncClient, admin_token: str):
    res = await client.get("/api/admin/users", headers={"Authorization": f"Bearer {admin_token}"})
    assert res.status_code == 200
    users = res.json()
    assert len(users) >= 1
    assert "username" in users[0]

@pytest.mark.asyncio
async def test_admin_users_list_student_forbidden(client: AsyncClient, student_token: str):
    res = await client.get("/api/admin/users", headers={"Authorization": f"Bearer {student_token}"})
    assert res.status_code == 403

@pytest.mark.asyncio
async def test_admin_stats_overview(client: AsyncClient, admin_token: str):
    res = await client.get("/api/admin/stats", headers={"Authorization": f"Bearer {admin_token}"})
    assert res.status_code == 200
    data = res.json()
    assert "total_fragments" in data
    assert "total_users" in data
    assert "confirmed_translations" in data

@pytest.mark.asyncio
async def test_export_translations_json(client: AsyncClient):
    res = await client.get("/api/export/translations?format=json")
    assert res.status_code == 200
    assert "application/json" in res.headers.get("content-type", "")
    data = res.json()
    assert "data" in data
    assert "total" in data

@pytest.mark.asyncio
async def test_export_translations_csv(client: AsyncClient):
    res = await client.get("/api/export/translations?format=csv")
    assert res.status_code == 200
    assert "text/csv" in res.headers.get("content-type", "")
    assert len(res.text) > 0

@pytest.mark.asyncio
async def test_corpus_analytics_stats(client: AsyncClient):
    res = await client.get("/api/stats/analytics")
    assert res.status_code == 200
    data = res.json()
    assert "summary" in data
    assert "total_users" in data["summary"]

@pytest.mark.asyncio
async def test_corpus_overview_stats(client: AsyncClient):
    res = await client.get("/api/stats/overview")
    assert res.status_code == 200
    data = res.json()
    assert isinstance(data, dict)

@pytest.mark.asyncio
async def test_gamification_level_threshold_boundaries():
    assert len(LEVEL_THRESHOLDS) == 10
    lvl1, t1, _ = get_level_info(0)
    assert lvl1 == 1 and t1 == "Начинающий"
    lvl10, t10, _ = get_level_info(15000)
    assert lvl10 == 10 and t10 == "Легенда языка"
