"""Test JFind scanner results endpoint."""

from datetime import datetime, timezone

import pytest
from httpx import AsyncClient
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import joinedload

from jfind_svc.db_model import ScanInfo
from jfind_svc.model import JavaRuntime, MetaInfo, ScannerResult


@pytest.fixture
def scanner_results() -> ScannerResult:
    """Create test scanner results data."""
    scan_time = datetime.now(timezone.utc).isoformat()
    return ScannerResult(
        meta=MetaInfo(
            scan_ts=scan_time,
            computer_name="test-computer",
            user_name="test-user",
            scan_duration="1s",
            has_oracle_jdk=True,
            count_result=2,
            count_require_license=1,
            scanned_dirs=10,
            scan_path="/test/path",
            platform_info="test-platform",
        ),
        runtimes=[
            JavaRuntime(
                java_executable="/usr/bin/java1",
                java_runtime="OpenJDK Runtime Environment",
                java_vendor="Oracle",
                is_oracle=True,
                java_version="1.8.0_292",
                java_version_major=8,
                java_version_update=292,
                require_license=True,
            ),
            JavaRuntime(
                java_executable="/usr/bin/java2",
                java_runtime="OpenJ9",
                java_vendor="Eclipse",
                is_oracle=False,
                java_version="11.0.3",
                java_version_major=11,
                java_version_update=3,
                require_license=False,
            ),
        ],
    )


@pytest.mark.asyncio
async def test_jfind_post_endpoint(
    test_client: AsyncClient,
    test_session: AsyncSession,
    scanner_results: ScannerResult,
):
    """Test POST /api/jfind endpoint for adding scan results."""
    # Send POST request
    response = await test_client.post("/api/jfind", json=scanner_results.model_dump())

    # Check response
    assert response.status_code == 200
    data = response.json()
    assert "result" in data
    assert data["result"] == "ok"
    assert "scan_id" in data
    scan_id = data["scan_id"]

    # Verify data in database
    stmt = select(ScanInfo).options(joinedload(ScanInfo.java_runtimes)).where(ScanInfo.id == scan_id)
    result = await test_session.execute(stmt)
    scan_info = result.unique().scalar_one()
    await test_session.refresh(scan_info)

    # Verify scan info
    assert scan_info.computer_name == "test-computer"
    assert scan_info.user_name == "test-user"
    assert scan_info.has_oracle_jdk is True
    assert scan_info.count_result == 2
    assert scan_info.count_require_license == 1
    assert scan_info.scanned_dirs == 10
    assert scan_info.scan_path == "/test/path"
    assert scan_info.most_recent is True  # Verify most_recent flag is set

    # Verify Java runtimes
    assert len(scan_info.java_runtimes) == 2

    java1 = next(j for j in scan_info.java_runtimes if j.java_executable == "/usr/bin/java1")
    assert java1.java_runtime == "OpenJDK Runtime Environment"
    assert java1.java_vendor == "Oracle"
    assert java1.is_oracle is True
    assert java1.java_version == "1.8.0_292"
    assert java1.java_version_major == 8
    assert java1.java_version_update == 292

    java2 = next(j for j in scan_info.java_runtimes if j.java_executable == "/usr/bin/java2")
    assert java2.java_runtime == "OpenJ9"
    assert java2.java_vendor == "Eclipse"
    assert java2.is_oracle is False
    assert java2.java_version == "11.0.3"
    assert java2.java_version_major == 11
    assert java2.java_version_update == 3
