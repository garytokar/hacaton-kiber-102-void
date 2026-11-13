import pytest
import uuid

@pytest.mark.asyncio
async def test_get_daily_fragment(client):
    res = await client.get("/api/fragments/daily")
    assert res.status_code == 200
    data = res.json()
    assert "text_ba" in data
    assert "translations" in data
    assert len(data["translations"]) > 0
    assert data["category"] is not None

@pytest.mark.asyncio
async def test_list_fragments_with_filters(client):
    res = await client.get("/api/fragments?limit=5")
    assert res.status_code == 200
    items = res.json()
    assert len(items) <= 5

@pytest.mark.asyncio
async def test_propose_translation_and_duplicate_vote(client):
    # Register fresh user to test voting and proposing cleanly
    uname = f"voter_{uuid.uuid4().hex[:6]}"
    reg = await client.post("/api/auth/register", json={
        "email": f"{uname}@mail.ru",
        "username": uname,
        "password": "password123"
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    # Propose new translation
    prop_res = await client.post("/api/translations", json={
        "fragment_id": 1,
        "translation_text": "Уникальный поэтический перевод на русский язык."
    }, headers=headers)

    assert prop_res.status_code == 200
    prop_data = prop_res.json()
    assert prop_data["success"] is True
    assert prop_data["xp_awarded"] == 8
    assert prop_data["currency_awarded"] == 5

    # Get fragment translations to find an unvoted translation
    frag_res = await client.get("/api/fragments/1", headers=headers)
    t_id = frag_res.json()["translations"][0]["id"]

    # Vote confirm
    vote_res = await client.post(f"/api/translations/{t_id}/vote", json={"vote_type": "confirm"}, headers=headers)
    assert vote_res.status_code == 200
    assert vote_res.json()["success"] is True

    # Vote duplicate -> MUST FAIL with 400
    dup_res = await client.post(f"/api/translations/{t_id}/vote", json={"vote_type": "confirm"}, headers=headers)
    assert dup_res.status_code == 400
    assert "уже проголосовали" in dup_res.json()["detail"]

@pytest.mark.asyncio
async def test_translation_history(client):
    res = await client.get("/api/translations/1/history")
    assert res.status_code == 200
    assert isinstance(res.json(), list)
