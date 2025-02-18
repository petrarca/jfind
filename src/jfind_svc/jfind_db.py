"""Database operations for JFind scanner results."""

from datetime import datetime
from typing import Optional

from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import joinedload

from jfind_svc.db_model import JavaInfo, ScanInfo
from jfind_svc.model import ScannerResults


async def save_scanner_results(session: AsyncSession, results: ScannerResults) -> ScanInfo:
    """Save scanner results to database."""
    # First, set most_recent=False for the current most recent record for this computer
    await session.execute(
        update(ScanInfo)
        .where(
            ScanInfo.computer_name == results.meta.computer_name,
            ScanInfo.most_recent == True
        )
        .values(most_recent=False)
    )

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
        scan_path=results.meta.scan_path,
        most_recent=True,  # Assumption is that records will be added
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
        .where(ScanInfo.most_recent == True)
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


async def get_scans_by_computer_name(session: AsyncSession, computer_name: str, limit: int = 0) -> list[ScanInfo]:
    """Get scans for a specific computer.

    Args:
        session: Database session
        computer_name: Name of the computer to get scans for
        limit: Maximum number of results to return. If -1, retrieve all records, if 0 retrieve only most_recent scan
            (most_recent = True), otherwise limit scans to limit.

    Returns:
        List of ScanInfo records with related JavaInfo records
    """
    query = (
        select(ScanInfo)
        .options(joinedload(ScanInfo.java_runtimes))  # Eagerly load relationships
        .where(ScanInfo.computer_name == computer_name)
    )

    if limit < 0:
        # retrieve all records
        pass
    elif limit == 0:
        # retrieve only most_recent scan (most_recent = True)
        query = query.where(ScanInfo.most_recent == True)
    else:
        # limit scans to limit
        query = query.order_by(ScanInfo.scan_ts.desc()).limit(limit)

    result = await session.execute(query)
    return list(result.unique().scalars().all())


async def get_oracle_jdks(session: AsyncSession, limit: int = 10) -> list[JavaInfo]:
    """Get all Oracle JDKs from the most recent scans.

    Args:
        session: Database session
        limit: Maximum number of results to return

    Returns:
        List of JavaInfo objects for Oracle JDKs
    """
    query = (
        select(JavaInfo)
        .options(joinedload(JavaInfo.scan))  # Eagerly load relationships
        .join(JavaInfo.scan)
        .where(JavaInfo.is_oracle == True)  # noqa: E712
        .where(ScanInfo.most_recent == True)
        .order_by(ScanInfo.scan_ts.desc())
        .limit(limit)
    )
    result = await session.execute(query)
    return list(result.scalars().all())


async def check_require_license(session: AsyncSession, computer_name: str) -> Optional[bool]:
    """Check if a computer has a JDK installed which require a license.

    Args:
        session: Database session
        computer_name: Name of computer to check

    Returns:
        True if the computer has as a commerical JDK installed, False if it doesn't, None if computer not found
    """
    # First check if we have any records for this computer
    stmt = select(JavaInfo).where(JavaInfo.computer_name == computer_name).limit(1)
    result = await session.execute(stmt)
    if result.first() is None:
        return None  # Computer not found

    # Check for Oracle JDKs on the computer
    stmt = (
        select(JavaInfo)
        .join(JavaInfo.scan)
        .where(
            JavaInfo.computer_name == computer_name,
            JavaInfo.require_license == True,  # noqa: E712
            ScanInfo.most_recent == True,
        )
        .limit(1)
    )
    result = await session.execute(stmt)
    return result.first() is not None
