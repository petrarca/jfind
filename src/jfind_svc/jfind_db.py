"""Database operations for JFind scanner results."""

from datetime import datetime
from typing import Optional

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import joinedload

from jfind_svc.db_model import JavaInfo, ScanInfo
from jfind_svc.model import ScannerResults


async def save_scanner_results(session: AsyncSession, results: ScannerResults) -> ScanInfo:
    """Save scanner results to the database.

    Args:
        session: Database session
        results: Scanner results from the API

    Returns:
        Created ScanInfo record
    """
    # Create scan info record
    scan_info = ScanInfo(
        scan_ts=datetime.fromisoformat(results.meta.scan_ts),
        computer_name=results.meta.computer_name,
        user_name=results.meta.user_name,
        scan_duration=results.meta.scan_duration,
        has_oracle_jdk=results.meta.has_oracle_jdk,
        count_result=results.meta.count_result,
        count_require_license=results.meta.count_require_license,
        scanned_dirs=results.meta.scanned_dirs,
    )
    session.add(scan_info)
    await session.flush()  # Get the scan_info.id

    # Create java info records
    for runtime in results.result:
        java_info = JavaInfo(
            scan_id=scan_info.id,
            computer_name=results.meta.computer_name,  # Add computer_name from scan metadata
            java_executable=runtime.java_executable,
            java_runtime=runtime.java_runtime,
            java_vendor=runtime.java_vendor,
            is_oracle=runtime.is_oracle,
            java_version=runtime.java_version,
            java_version_major=runtime.java_version_major,
            java_version_update=runtime.java_version_update,
        )
        session.add(java_info)

    await session.commit()
    await session.refresh(scan_info, ["java_runtimes"])  # Refresh relationships
    return scan_info


async def get_latest_scans(session: AsyncSession, limit: int = 10) -> list[ScanInfo]:
    """Get the latest scans with their Java runtime information.

    Args:
        session: Database session
        limit: Maximum number of scans to return

    Returns:
        List of ScanInfo records with related JavaInfo records
    """
    query = (
        select(ScanInfo)
        .options(joinedload(ScanInfo.java_runtimes))  # Eagerly load relationships
        .order_by(ScanInfo.scan_ts.desc())
        .limit(limit)
    )
    result = await session.execute(query)
    return list(result.unique().scalars().all())


async def get_scan_by_id(session: AsyncSession, scan_id: int) -> Optional[ScanInfo]:
    """Get a scan by its ID.

    Args:
        session: Database session
        scan_id: ID of the scan to retrieve

    Returns:
        ScanInfo record if found, None otherwise
    """
    query = (
        select(ScanInfo)
        .options(joinedload(ScanInfo.java_runtimes))  # Eagerly load relationships
        .where(ScanInfo.id == scan_id)
    )
    result = await session.execute(query)
    return result.unique().scalar_one_or_none()


async def get_scans_by_computer_name(session: AsyncSession, computer_name: str, limit: int = 10) -> list[ScanInfo]:
    """Get scans for a specific computer.

    Args:
        session: Database session
        computer_name: Name of the computer to get scans for
        limit: Maximum number of results to return

    Returns:
        List of ScanInfo records with related JavaInfo records
    """
    query = (
        select(ScanInfo)
        .options(joinedload(ScanInfo.java_runtimes))  # Eagerly load relationships
        .where(ScanInfo.computer_name == computer_name)
        .order_by(ScanInfo.scan_ts.desc())
        .limit(limit)
    )
    result = await session.execute(query)
    return list(result.unique().scalars().all())


async def get_oracle_jdks(session: AsyncSession, limit: int = 10) -> list[JavaInfo]:
    """Get all Oracle JDKs from the database.

    Args:
        session: Database session
        limit: Maximum number of results to return

    Returns:
        List of JavaInfo objects for Oracle JDKs
    """
    stmt = (
        select(JavaInfo)
        .where(JavaInfo.is_oracle == True)  # noqa: E712
        .order_by(JavaInfo.id.desc())
        .limit(limit)
    )
    result = await session.execute(stmt)
    return list(result.scalars().all())


async def has_oracle_jdk(session: AsyncSession, computer_name: str) -> Optional[bool]:
    """Check if a computer has Oracle JDK installed.

    Args:
        session: Database session
        computer_name: Name of computer to check

    Returns:
        True if the computer has Oracle JDK, False if it doesn't, None if computer not found
    """
    # First check if we have any records for this computer
    stmt = (
        select(JavaInfo)
        .where(JavaInfo.computer_name == computer_name)
        .limit(1)
    )
    result = await session.execute(stmt)
    if result.first() is None:
        return None  # Computer not found
    
    # Check for Oracle JDKs on the computer
    stmt = (
        select(JavaInfo)
        .where(
            JavaInfo.computer_name == computer_name,
            JavaInfo.is_oracle == True,  # noqa: E712
        )
        .limit(1)
    )
    result = await session.execute(stmt)
    return result.first() is not None
