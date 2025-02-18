"""Database configuration and session management."""

import os
from pathlib import Path

from dotenv import load_dotenv
from sqlalchemy.ext.asyncio import AsyncEngine, AsyncSession, async_sessionmaker, create_async_engine

from jfind_svc.db_model import Base

# Load environment variables from .env files
load_dotenv()  # Load from .env in current directory
load_dotenv(Path.home() / ".env")  # Load from ~/.env

# Default database URLs
DEFAULT_SQLITE_URL = "sqlite+aiosqlite:///./jfind.db"
DEFAULT_POSTGRES_URL = "postgresql+asyncpg://postgres:postgres@localhost:5432/jfind"


def get_database_url() -> str:
    """Get database URL based on environment and configuration.

    Priority:
    1. Command line argument (set via DATABASE_URL environment variable in main.py)
    2. Environment variable (from shell or .env files)
    3. Default value based on environment
    """
    # Check for explicit DATABASE_URL (from env, .env files, or command line)
    if database_url := os.getenv("DATABASE_URL"):
        return database_url

    # Use default based on environment
    is_production = os.getenv("ENV", "development") == "production"
    return DEFAULT_POSTGRES_URL if is_production else DEFAULT_SQLITE_URL


# Create async engine and session factory
engine: AsyncEngine = create_async_engine(
    get_database_url(),
    echo=True,
    # Required for SQLite
    connect_args={"check_same_thread": False} if "sqlite" in get_database_url() else {},
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
