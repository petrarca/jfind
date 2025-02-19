"""Test JFind scanner results GET endpoint."""

from datetime import datetime, timezone

import pytest
import pytest_asyncio
from httpx import AsyncClient
from sqlalchemy.ext.asyncio import AsyncSession

from jfind_svc.db_model import JavaInfo, ScanInfo


@pytest_asyncio.fixture
async def test_scan_data(test_session: AsyncSession) -> ScanInfo:
    """Create test scan data in the database."""
    # Create scan info
    scan_info = ScanInfo(
        scan_ts=datetime.now(timezone.utc),
        computer_name="test-computer-1",
        user_name="test-user-1",
        scan_duration="2s",
        has_oracle_jdk=True,
        count_result=3,
        count_require_license=2,
        scanned_dirs=15,
        scan_path="/test/path/1",
        platform_info="test-platform-1",
        most_recent=True,
    )
    test_session.add(scan_info)
    await test_session.flush()  # Get the scan_info.id

    # Create java info records
    java_infos = [
        JavaInfo(
            scan_id=scan_info.id,
            computer_name="test-computer-1",
            java_executable="/usr/bin/java1",
            java_runtime="OpenJDK Runtime Environment",
            java_vendor="Oracle",
            is_oracle=True,
            java_version="1.8.0_292",
            java_version_major=8,
            java_version_update=292,
            require_license=True,
        ),
        JavaInfo(
            scan_id=scan_info.id,
            computer_name="test-computer-1",
            java_executable="/usr/bin/java2",
            java_runtime="OpenJ9",
            java_vendor="Eclipse",
            is_oracle=False,
            java_version="11.0.3",
            java_version_major=11,
            java_version_update=3,
            require_license=False,
        ),
        JavaInfo(
            scan_id=scan_info.id,
            computer_name="test-computer-1",
            java_executable="/usr/bin/java3",
            java_runtime="OpenJDK Runtime Environment",
            java_vendor="AdoptOpenJDK",
            is_oracle=False,
            java_version="17.0.1",
            java_version_major=17,
            java_version_update=1,
            require_license=False,
        ),
    ]
    for java_info in java_infos:
        test_session.add(java_info)

    await test_session.commit()
    await test_session.refresh(scan_info)
    return scan_info


@pytest.mark.asyncio
async def test_jfind_get_endpoint(
    test_client: AsyncClient,
    test_scan_data: ScanInfo,
):
    """Test GET /api/jfind endpoint for retrieving scan results."""
    # Send GET request
    response = await test_client.get(f"/api/jfind?scan_id={test_scan_data.id}")

    # Check response
    assert response.status_code == 200
    data = response.json()
    assert isinstance(data, list)
    assert len(data) == 1
    scan_data = data[0]

    # Verify scan info
    meta = scan_data["meta"]
    assert meta["scan_id"] == test_scan_data.id
    assert meta["computer_name"] == "test-computer-1"
    assert meta["user_name"] == "test-user-1"
    assert meta["has_oracle_jdk"] is True
    assert meta["count_result"] == 3
    assert meta["count_require_license"] == 2
    assert meta["scanned_dirs"] == 15
    assert meta["scan_path"] == "/test/path/1"
    assert meta["platform_info"] == "test-platform-1"

    # Verify Java runtimes
    runtimes = scan_data["runtimes"]
    assert len(runtimes) == 3

    # Verify each Java runtime
    java1 = next(j for j in runtimes if j["java_executable"] == "/usr/bin/java1")
    assert java1["java_runtime"] == "OpenJDK Runtime Environment"
    assert java1["java_vendor"] == "Oracle"
    assert java1["is_oracle"] is True
    assert java1["java_version"] == "1.8.0_292"
    assert java1["java_version_major"] == 8
    assert java1["java_version_update"] == 292

    java2 = next(j for j in runtimes if j["java_executable"] == "/usr/bin/java2")
    assert java2["java_runtime"] == "OpenJ9"
    assert java2["java_vendor"] == "Eclipse"
    assert java2["is_oracle"] is False
    assert java2["java_version"] == "11.0.3"
    assert java2["java_version_major"] == 11
    assert java2["java_version_update"] == 3

    java3 = next(j for j in runtimes if j["java_executable"] == "/usr/bin/java3")
    assert java3["java_runtime"] == "OpenJDK Runtime Environment"
    assert java3["java_vendor"] == "AdoptOpenJDK"
    assert java3["is_oracle"] is False
    assert java3["java_version"] == "17.0.1"
    assert java3["java_version_major"] == 17
    assert java3["java_version_update"] == 1


@pytest.mark.asyncio
async def test_jfind_get_endpoint_not_found(test_client: AsyncClient):
    """Test GET /api/jfind endpoint with non-existent scan_id."""
    response = await test_client.get("/api/jfind?scan_id=999999")
    assert response.status_code == 404
    data = response.json()
    assert data["detail"] == "Scan with ID 999999 not found"
