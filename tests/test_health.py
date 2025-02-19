"""Test health check endpoint."""

import os
from datetime import datetime

import pytest
from httpx import AsyncClient


@pytest.mark.asyncio
async def test_health_endpoint(test_client: AsyncClient):
    """Test health check endpoint returns expected data structure."""
    response = await test_client.get("/health")

    # Check status code
    assert response.status_code == 200

    # Check response structure
    data = response.json()
    assert "hostname" in data
    assert "process_id" in data
    assert "timestamp" in data

    # Validate timestamp format
    timestamp = datetime.fromisoformat(data["timestamp"])
    assert isinstance(timestamp, datetime)

    # Validate process_id is an integer
    assert isinstance(data["process_id"], int)
    assert data["process_id"] == os.getpid()

    # Validate hostname is a non-empty string
    assert isinstance(data["hostname"], str)
    assert len(data["hostname"]) > 0
