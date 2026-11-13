import pytest
from app.services.consensus import (
    canonicalize, levenshtein_distance, levenshtein_similarity,
    trigram_similarity, compute_combined_similarity
)

def test_fuzzy_canonicalization():
    text = "«Ат даны — яуҙа беленер!»"
    canon = canonicalize(text)
    assert '«' not in canon
    assert '»' not in canon
    assert '—' not in canon
    assert '-' in canon
    assert canon == 'ат даны - яуҙа беленер'

def test_levenshtein_and_trigram():
    s1 = "Слава джигита в бою"
    s2 = "Слава джигита в бою."
    sim = levenshtein_similarity(s1, s2)
    assert sim > 0.9

    s3 = "Слава коня познаётся в бою"
    s4 = "Слава коня познается в бою"
    combined = compute_combined_similarity(s3, s4)
    assert combined >= 0.85

@pytest.mark.anyio
async def test_lang_pair_endpoints(client, student_token):
    headers = {"Authorization": f"Bearer {student_token}"}
    
    # 1. Get initial lang pair
    res = await client.get("/api/users/me/lang-pair", headers=headers)
    assert res.status_code == 200
    data = res.json()
    assert "preferred_lang_pair" in data
    assert "available_pairs" in data

    # 2. Update lang pair to ba-kk
    res_put = await client.put("/api/users/me/lang-pair", json={"lang_pair": "ba-kk"}, headers=headers)
    assert res_put.status_code == 200
    assert res_put.json()["preferred_lang_pair"] == "ba-kk"

@pytest.mark.anyio
async def test_leaderboard_classes_and_vocabulary(client):
    # Classes leaderboard
    res = await client.get("/api/leaderboard/classes")
    assert res.status_code == 200
    assert isinstance(res.json(), list)

    # Vocabulary leaderboard
    res_voc = await client.get("/api/leaderboard/vocabulary")
    assert res_voc.status_code == 200
    assert isinstance(res_voc.json(), list)

@pytest.mark.anyio
async def test_lms_classes_flow(client, student_token):
    headers = {"Authorization": f"Bearer {student_token}"}

    # 1. Get schools
    res_schools = await client.get("/api/lms/schools", headers=headers)
    assert res_schools.status_code == 200
    schools = res_schools.json()
    assert len(schools) > 0
    school_id = schools[0]["id"]

    # 2. Create class
    res_create = await client.post(
        "/api/lms/classes",
        json={"school_id": school_id, "name": "10-В Тестовый", "grade_number": 10, "letter": "В"},
        headers=headers
    )
    assert res_create.status_code == 200
    class_data = res_create.json()
    invite_code = class_data["invite_code"]
    assert invite_code.startswith("INV-")

    # 3. Join class with invite code
    res_join = await client.post(
        "/api/lms/classes/join",
        json={"invite_code": invite_code},
        headers=headers
    )
    assert res_join.status_code == 200
    assert res_join.json()["success"] is True
