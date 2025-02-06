"""Database configuration and session management."""

import os

from sqlalchemy.ext.asyncio import AsyncEngine, AsyncSession, async_sessionmaker, create_async_engine

from jfind_svc.db_model import Base

# Database URLs
SQLITE_URL = "sqlite+aiosqlite:///./jfind.db"
POSTGRES_URL = os.getenv("DATABASE_URL", "postgresql+asyncpg://postgres:postgres@localhost:5432/jfind")

# Use SQLite for development, PostgreSQL for production
is_production = os.getenv("ENV", "development") == "production"
DATABASE_URL = POSTGRES_URL if is_production else SQLITE_URL


# Create async engine and session factory
engine: AsyncEngine = create_async_engine(
    DATABASE_URL,
    echo=True,
    # Required for SQLite
    connect_args={"check_same_thread": False} if not is_production else {},
)

async_session = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,  # Don't expire objects after commit
)


async def init_db():
    """Initialize the database."""
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)


async def get_session() -> AsyncSession:
    """Get a database session."""
    async with async_session() as session:
        yield session
