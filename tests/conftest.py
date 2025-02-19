"""Shared test fixtures and utilities."""

from typing import AsyncGenerator, Callable

import pytest_asyncio
from httpx import ASGITransport, AsyncClient
from sqlalchemy.ext.asyncio import AsyncEngine, AsyncSession, async_sessionmaker, create_async_engine

from jfind_svc.db import get_session
from jfind_svc.db_model import Base
from jfind_svc.main import app

# Test database URL for in-memory SQLite
TEST_DATABASE_URL = "sqlite+aiosqlite:///:memory:"


@pytest_asyncio.fixture
async def test_engine() -> AsyncEngine:
    """Create a test database engine."""
    engine = create_async_engine(TEST_DATABASE_URL, echo=True, connect_args={"check_same_thread": False})

    # Create all tables
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    yield engine

    # Drop all tables after test
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)


@pytest_asyncio.fixture
async def test_session_maker(test_engine: AsyncEngine) -> async_sessionmaker[AsyncSession]:
    """Create session maker for tests."""
    session_maker = async_sessionmaker(test_engine, class_=AsyncSession, expire_on_commit=False)
    return session_maker


@pytest_asyncio.fixture
async def test_session(test_session_maker: async_sessionmaker[AsyncSession]) -> AsyncGenerator[AsyncSession, None]:
    """Create test database session."""
    async with test_session_maker() as session:
        yield session


def get_test_session_dependency(
    session_maker: async_sessionmaker[AsyncSession],
) -> Callable[[], AsyncGenerator[AsyncSession, None]]:
    """Create a test session dependency."""

    async def get_session() -> AsyncGenerator[AsyncSession, None]:
        async with session_maker() as session:
            yield session

    return get_session


@pytest_asyncio.fixture
async def test_client(test_session_maker: async_sessionmaker[AsyncSession]) -> AsyncGenerator[AsyncClient, None]:
    """Create a test client with the test database session."""
    # Override the get_session dependency
    app.dependency_overrides[get_session] = get_test_session_dependency(test_session_maker)

    # Create test client with ASGI transport
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        yield client

    # Clean up
    app.dependency_overrides.clear()
