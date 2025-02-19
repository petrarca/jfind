"""Database models for the JFind service."""

from datetime import datetime, timezone
from typing import Optional

from sqlalchemy import ForeignKey, String
from sqlalchemy.ext.asyncio import AsyncAttrs
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship


class Base(AsyncAttrs, DeclarativeBase):
    """Base class for all database models."""

    pass


class ScanInfo(Base):
    """Database model for scan metadata."""

    __tablename__ = "scan_info"

    id: Mapped[int] = mapped_column(primary_key=True)
    scan_ts: Mapped[datetime] = mapped_column()
    computer_name: Mapped[str] = mapped_column(String(255))
    user_name: Mapped[str] = mapped_column(String(255))
    scan_duration: Mapped[str] = mapped_column(String(50))
    has_oracle_jdk: Mapped[bool] = mapped_column()
    count_result: Mapped[int] = mapped_column()
    count_require_license: Mapped[int] = mapped_column(nullable=True)
    scanned_dirs: Mapped[int] = mapped_column()
    scan_path: Mapped[str] = mapped_column(String(1024), nullable=True)
    most_recent: Mapped[bool] = mapped_column(nullable=True)
    platform_info: Mapped[str] = mapped_column(String(255), nullable=True)
    created_at: Mapped[datetime] = mapped_column(default=lambda: datetime.now(timezone.utc))

    # Relationship to JavaInfo
    java_runtimes: Mapped[list["JavaInfo"]] = relationship(back_populates="scan", cascade="all, delete-orphan")


class JavaInfo(Base):
    """Database model for Java runtime information."""

    __tablename__ = "java_info"

    id: Mapped[int] = mapped_column(primary_key=True)
    scan_id: Mapped[int] = mapped_column(ForeignKey("scan_info.id"))
    computer_name: Mapped[str] = mapped_column(String(255))
    java_executable: Mapped[str] = mapped_column(String(1024))
    java_runtime: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    java_vendor: Mapped[Optional[str]] = mapped_column(String(255), nullable=True)
    is_oracle: Mapped[Optional[bool]] = mapped_column(nullable=True)
    java_version: Mapped[Optional[str]] = mapped_column(String(50), nullable=True)
    java_version_major: Mapped[Optional[int]] = mapped_column(nullable=True)
    java_version_update: Mapped[Optional[int]] = mapped_column(nullable=True)
    require_license: Mapped[Optional[bool]] = mapped_column(nullable=True)
    created_at: Mapped[datetime] = mapped_column(default=lambda: datetime.now(timezone.utc))

    # Relationship to ScanInfo
    scan: Mapped[ScanInfo] = relationship(back_populates="java_runtimes")
