import pytest
import uuid

@pytest.mark.asyncio
async def test_register_new_user(client):
    unique_email = f"user_{uuid.uuid4().hex[:8]}@test.ru"
    unique_name = f"user_{uuid.uuid4().hex[:8]}"

    payload = {
        "email": unique_email,
        "username": unique_name,
        "password": "strongpassword123",
        "full_name": "Тестовый Хранитель",
        "role": "student"
    }

    res = await client.post("/api/auth/register", json=payload)
    assert res.status_code == 200, res.text
    data = res.json()
    assert "access_token" in data
    assert data["user"]["email"] == unique_email
    assert data["user"]["stats"]["level"] == 1
    assert data["user"]["stats"]["currency"] == 50

@pytest.mark.asyncio
async def test_register_duplicate_email_rejected(client):
    unique_email = f"user_{uuid.uuid4().hex[:8]}@test.ru"
    payload = {
        "email": unique_email,
        "username": f"user_{uuid.uuid4().hex[:8]}",
        "password": "strongpassword123",
    }
    res1 = await client.post("/api/auth/register", json=payload)
    assert res1.status_code == 200

    # Attempt second registration with same email
    payload2 = {
        "email": unique_email,
        "username": f"another_{uuid.uuid4().hex[:8]}",
        "password": "strongpassword123",
    }
    res2 = await client.post("/api/auth/register", json=payload2)
    assert res2.status_code == 400
    assert "уже существует" in res2.json()["detail"]

@pytest.mark.asyncio
async def test_login_flow(client):
    # Valid login with demo student
    login_payload = {
        "email": "student_timur@yanlander.ru",
        "password": "password123"
    }
    res = await client.post("/api/auth/login", json=login_payload)
    assert res.status_code == 200
    data = res.json()
    assert "access_token" in data
    assert data["user"]["username"] == "timur_batyr"

    # Invalid password login
    bad_payload = {
        "email": "student_timur@yanlander.ru",
        "password": "wrong_password_xyz"
    }
    bad_res = await client.post("/api/auth/login", json=bad_payload)
    assert bad_res.status_code == 400

@pytest.mark.asyncio
async def test_get_current_user(client, student_token):
    # Unauthorized
    unauth_res = await client.get("/api/auth/me")
    assert unauth_res.status_code == 401

    # Authorized
    auth_res = await client.get("/api/auth/me", headers={"Authorization": f"Bearer {student_token}"})
    assert auth_res.status_code == 200
    assert auth_res.json()["role"] == "student"
