"""JFind scanner results endpoint."""

from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.responses import JSONResponse
from sqlalchemy.ext.asyncio import AsyncSession

from jfind_svc.db import get_session
from jfind_svc.jfind_db import (
    ScanInfo,
    get_latest_scans,
    get_oracle_jdks,
    get_scan_by_id,
    get_scans_by_computer_name,
    save_scanner_results,
)
from jfind_svc.model import ScannerResults

router = APIRouter(tags=["jfind"])

# Get database session dependency
db_session = Depends(get_session)


@router.post("/jfind", status_code=status.HTTP_200_OK)
async def process_scanner_results(results: ScannerResults, session: AsyncSession = db_session) -> JSONResponse:
    """Process results from the jfind scanner.

    Returns:
        200 OK with {"result": "ok", "scan_id": <id>} if data is valid
        422 Unprocessable Entity if data validation fails
    """
    # Save results to database
    scan_info = await save_scanner_results(session, results)

    # Log success
    print(f"Saved scan from {scan_info.computer_name} with {scan_info.count_result} Java runtimes")

    return JSONResponse(content={"result": "ok", "scan_id": scan_info.id}, status_code=status.HTTP_200_OK)


@router.get("/jfind/scans", status_code=status.HTTP_200_OK)
async def get_scans(limit: int = 10, session: AsyncSession = db_session) -> JSONResponse:
    """Get the latest scan results.

    Args:
        limit: Maximum number of scans to return (default: 10)
        session: Database session

    Returns:
        200 OK with list of scans and their Java runtime information
    """
    scans = await get_latest_scans(session, limit)
    response = [_format_scan_response(scan) for scan in scans]
    return JSONResponse(content=response, status_code=status.HTTP_200_OK)


@router.get("/jfind", status_code=status.HTTP_200_OK)
async def query_scans(
    computer_name: Optional[str] = None,
    scan_id: Optional[int] = None,
    limit: int = 10,
    session: AsyncSession = db_session,
) -> JSONResponse:
    """Query scan results by computer name or scan ID.

    Args:
        computer_name: Optional name of computer to query
        scan_id: Optional scan ID to query
        limit: Maximum number of results to return (default: 10)
        session: Database session

    Returns:
        200 OK with matching scan results
        404 Not Found if scan_id is specified but not found
    """
    if scan_id is not None:
        # Query by scan ID
        scan = await get_scan_by_id(session, scan_id)
        if scan is None:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=f"Scan with ID {scan_id} not found")
        response = [_format_scan_response(scan)]
    elif computer_name is not None:
        # Query by computer name
        scans = await get_scans_by_computer_name(session, computer_name, limit)
        response = [_format_scan_response(scan) for scan in scans]
    else:
        # No query parameters, return latest scans
        scans = await get_latest_scans(session, limit)
        response = [_format_scan_response(scan) for scan in scans]

    return JSONResponse(content=response, status_code=status.HTTP_200_OK)


@router.get("/jfind/computer/{computer_name}", status_code=status.HTTP_200_OK)
async def get_scans_by_computer(computer_name: str, limit: int = 10, session: AsyncSession = db_session) -> JSONResponse:
    """Get scan results for a specific computer.

    Args:
        computer_name: Name of computer to get scans for
        limit: Maximum number of scans to return (default: 10)
        session: Database session

    Returns:
        200 OK with matching scan results
    """
    scans = await get_scans_by_computer_name(session, computer_name, limit)
    response = [_format_scan_response(scan) for scan in scans]
    return JSONResponse(content=response, status_code=status.HTTP_200_OK)


@router.get("/jfind/jdk/oracle", status_code=status.HTTP_200_OK)
async def get_oracle_java_runtimes(limit: int = 10, session: AsyncSession = db_session) -> JSONResponse:
    """Get all Oracle Java runtimes.

    Args:
        limit: Maximum number of results to return (default: 10)
        session: Database session

    Returns:
        200 OK with list of Oracle Java runtimes
    """
    java_infos = await get_oracle_jdks(session, limit)
    response = [
        {
            "scan_id": java.scan_id,
            "computer_name": java.computer_name,
            "java_executable": java.java_executable,
            "java_runtime": java.java_runtime,
            "java_vendor": java.java_vendor,
            "is_oracle": java.is_oracle,
            "java_version": java.java_version,
            "java_version_major": java.java_version_major,
            "java_version_update": java.java_version_update,
        }
        for java in java_infos
    ]
    return JSONResponse(content=response, status_code=status.HTTP_200_OK)


def _format_scan_response(scan: ScanInfo) -> dict:
    """Format a single scan result for API response."""
    return {
        "meta": {
            "scan_id": scan.id,
            "scan_ts": scan.scan_ts.isoformat(),
            "computer_name": scan.computer_name,
            "user_name": scan.user_name,
            "scan_duration": scan.scan_duration,
            "has_oracle_jdk": scan.has_oracle_jdk,
            "count_result": scan.count_result,
            "scanned_dirs": scan.scanned_dirs,
        },
        "result": [
            {
                "java_executable": java.java_executable,
                "java_runtime": java.java_runtime,
                "java_vendor": java.java_vendor,
                "is_oracle": java.is_oracle,
                "java_version": java.java_version,
                "java_version_major": java.java_version_major,
                "java_version_update": java.java_version_update,
            }
            for java in scan.java_runtimes
        ],
    }
