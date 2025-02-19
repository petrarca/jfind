"""JFind scanner results endpoint."""

from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.responses import JSONResponse
from sqlalchemy.ext.asyncio import AsyncSession

from jfind_svc.db import get_session
from jfind_svc.jfind_db import (
    JavaInfo,
    ScanInfo,
    check_require_license,
    get_latest_scans,
    get_oracle_jdks,
    get_scan_by_id,
    get_scans_by_computer_name,
    save_scanner_results,
)
from jfind_svc.model import JavaRuntime, MetaInfo, ScannerResult

router = APIRouter(tags=["jfind"])

# Get database session dependency
db_session = Depends(get_session)


@router.post("/jfind", status_code=status.HTTP_200_OK)
async def process_scanner_results(results: ScannerResult, session: AsyncSession = db_session) -> JSONResponse:
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
    response = [_format_scan_response_data(scan) for scan in scans]
    return JSONResponse(content=response, status_code=status.HTTP_200_OK)


@router.get("/jfind/scans/{computer_name}", status_code=status.HTTP_200_OK)
async def get_scans_by_computer(computer_name: str, limit: int = 0, session: AsyncSession = db_session) -> JSONResponse:
    """Get scan results for a specific computer.

    Args:
        computer_name: Name of computer to get scans for
        limit: Maximum number of scans to return, if 0 retrieve only most recent scan. If < 0 retrieve all scans (default: 0)
        session: Database session

    Returns:
        200 OK with matching scan results
    """
    scans = await get_scans_by_computer_name(session, computer_name, limit)
    response = [_format_scan_response(scan) for scan in scans]
    return JSONResponse(content=response, status_code=status.HTTP_200_OK)


@router.get("/jfind/oracle", status_code=status.HTTP_200_OK)
async def get_oracle_java_runtimes(limit: int = 10, session: AsyncSession = db_session) -> JSONResponse:
    """Get all Oracle Java runtimes.

    Args:
        limit: Maximum number of results to return (default: 10)
        session: Database session

    Returns:
        200 OK with list of Oracle Java runtimes
    """
    java_infos = await get_oracle_jdks(session, limit)
    response = [_format_java_response_data(java_info) for java_info in java_infos]

    return JSONResponse(content=response, status_code=status.HTTP_200_OK)


@router.get("/jfind/require_license/{computer_name}", status_code=status.HTTP_200_OK)
async def check_oracle_jdk(computer_name: str, session: AsyncSession = db_session) -> JSONResponse:
    """Check if a computer has one or more JDKs installed which require a license

    Args:
        computer_name: Name of computer to check
        session: Database session

    Returns:
        200 OK with {
            "computer_name": str,
            "require_license": "true"/"false"/"unknown"
        }
        - "true": Computer has Oracle JDK installed
        - "false": Computer has Java records but no Oracle JDK
        - "unknown": No records found for this computer
    """
    require_license = await check_require_license(session, computer_name)

    # Convert boolean/None to true/false/unknown
    result = {True: "true", False: "false", None: "unknown"}[require_license]

    return JSONResponse(content={"computer_name": computer_name, "require_license": result}, status_code=status.HTTP_200_OK)


def _format_scan_response(scan: ScanInfo) -> dict[str, any]:
    """Format a single scan result"""
    return {
        "meta": _format_scan_response_data(scan),
        "runtimes": [_format_java_response_data(runtime) for runtime in scan.java_runtimes],
    }


def _format_scan_response_data(scan: ScanInfo) -> dict[str, any]:
    """Format the scan data"""
    return MetaInfo.model_validate(scan).model_dump()


def _format_java_response_data(java: JavaInfo) -> dict[str, any]:
    """Format a runtime record"""
    return JavaRuntime.model_validate(java).model_dump()
