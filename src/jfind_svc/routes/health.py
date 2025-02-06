"""Health check endpoint."""

import os
import socket
from datetime import datetime

from fastapi import APIRouter

router = APIRouter(tags=["health"])


@router.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"hostname": socket.gethostname(), "process_id": os.getpid(), "timestamp": datetime.now().isoformat()}
