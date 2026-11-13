import pytest

@pytest.mark.asyncio
async def test_admin_permissions(client, student_token, admin_token, expert_token):
    # Student must get 403 Forbidden on /api/admin/review
    forbidden_res = await client.get("/api/admin/review", headers={"Authorization": f"Bearer {student_token}"})
    assert forbidden_res.status_code == 403

    # Expert can access review queue
    expert_res = await client.get("/api/admin/review", headers={"Authorization": f"Bearer {expert_token}"})
    assert expert_res.status_code == 200
    assert isinstance(expert_res.json(), list)

    # Admin can access users list
    admin_users_res = await client.get("/api/admin/users", headers={"Authorization": f"Bearer {admin_token}"})
    assert admin_users_res.status_code == 200
    assert len(admin_users_res.json()) > 0

@pytest.mark.asyncio
async def test_admin_stats(client, admin_token):
    res = await client.get("/api/admin/stats", headers={"Authorization": f"Bearer {admin_token}"})
    assert res.status_code == 200
    data = res.json()
    assert data["total_fragments"] >= 30
    assert data["total_users"] >= 10
    assert data["confirmed_translations"] >= 20

@pytest.mark.asyncio
async def test_export_corpus_formats(client):
    # JSON export
    json_res = await client.get("/api/export/translations?format=json")
    assert json_res.status_code == 200
    assert "data" in json_res.json()
    assert json_res.json()["total"] > 0

    # CSV export
    csv_res = await client.get("/api/export/translations?format=csv")
    assert csv_res.status_code == 200
    assert "text/csv" in csv_res.headers["content-type"]
    assert "text_ba" in csv_res.text
    assert "translation_ru" in csv_res.text
