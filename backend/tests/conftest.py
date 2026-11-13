import pytest
import asyncio
from httpx import AsyncClient, ASGITransport
from app.main import app, lifespan
from app.core.database import engine, Base, AsyncSessionLocal
from app.seed.demo_data import seed_all
from app.core.security import create_access_token
from app.models.models import User
from sqlalchemy import select

@pytest.fixture(scope="session", autouse=True)
async def init_db():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    async with AsyncSessionLocal() as db:
        await seed_all(db)
    yield

@pytest.fixture
async def client():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        yield ac

@pytest.fixture
async def admin_token():
    async with AsyncSessionLocal() as db:
        res = await db.execute(select(User).where(User.role == "admin"))
        user = res.scalars().first()
        assert user is not None
        return create_access_token(user.id)

@pytest.fixture
async def expert_token():
    async with AsyncSessionLocal() as db:
        res = await db.execute(select(User).where(User.role == "expert"))
        user = res.scalars().first()
        assert user is not None
        return create_access_token(user.id)

@pytest.fixture
async def student_token():
    async with AsyncSessionLocal() as db:
        res = await db.execute(select(User).where(User.role == "student"))
        user = res.scalars().first()
        assert user is not None
        return create_access_token(user.id)
