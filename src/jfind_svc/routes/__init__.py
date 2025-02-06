"""Routes package."""

from fastapi import APIRouter

from .health import router as health_router
from .jfind import router as jfind_router

# Create main router without prefix
router = APIRouter()

# Create health router without prefix
health = APIRouter()
health.include_router(health_router)

# Include jfind router in the API router
router.include_router(jfind_router)

# Export both routers
__all__ = ["router", "health"]
