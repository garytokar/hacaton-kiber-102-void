import pytest
import uuid

@pytest.mark.asyncio
async def test_treasury_items(client):
    res = await client.get("/api/stats/treasury")
    assert res.status_code == 200
    data = res.json()
    assert "categories" in data
    assert "items" in data
    assert len(data["items"]) > 0
    # Every item in treasury MUST be confirmed
    first_item = data["items"][0]
    assert "text_ba" in first_item
    assert "translation_text" in first_item

@pytest.mark.asyncio
async def test_dictionary_lifecycle(client):
    uname = f"dict_user_{uuid.uuid4().hex[:6]}"
    reg = await client.post("/api/auth/register", json={
        "email": f"{uname}@mail.ru",
        "username": uname,
        "password": "password123"
    })
    token = reg.json()["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    # Add word
    add_res = await client.post("/api/dictionary", json={
        "word": "Ҡурай",
        "translation": "Курай (духовой инструмент)",
        "context": "Башкирский национальный инструмент"
    }, headers=headers)
    assert add_res.status_code == 200
    entry_id = add_res.json()["id"]
    assert add_res.json()["word"] == "Ҡурай"

    # List dictionary
    list_res = await client.get("/api/dictionary", headers=headers)
    assert list_res.status_code == 200
    assert len(list_res.json()) >= 1

    # Toggle favorite
    fav_res = await client.patch(f"/api/dictionary/{entry_id}/favorite", headers=headers)
    assert fav_res.status_code == 200
    assert fav_res.json()["is_favorite"] is True

    # Delete word
    del_res = await client.delete(f"/api/dictionary/{entry_id}", headers=headers)
    assert del_res.status_code == 200
    assert del_res.json()["success"] is True
