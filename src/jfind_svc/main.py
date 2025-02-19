"""JFind Service main module."""

import argparse
import os
import sys
from contextlib import asynccontextmanager
from typing import NamedTuple, Optional

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from loguru import logger

from jfind_svc.db import init_db
from jfind_svc.routes.health import router as health_router
from jfind_svc.routes.jfind import router as jfind_router


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """Lifespan events for the FastAPI application.

    This handles startup and shutdown events:
    - Startup: Initialize the database
    - Shutdown: Clean up resources (if needed)
    """
    # Startup: Initialize the database
    await init_db()
    yield
    # Shutdown: Clean up if needed
    # Currently no cleanup needed


app = FastAPI(
    title="JFind Service",
    description="Backend service for JFind",
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
    lifespan=lifespan,
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(health_router)  # /health endpoint
app.include_router(jfind_router, prefix="/api")  # /api/jfind endpoints


class ServerConfig(NamedTuple):
    """Server configuration."""

    host: str = "0.0.0.0"
    port: int = 8000
    database_url: Optional[str] = None


def parse_args() -> ServerConfig:
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(description="JFind Service")
    parser.add_argument("--host", default="0.0.0.0", help="Host to bind to")
    parser.add_argument("--port", type=int, default=8000, help="Port to bind to")
    parser.add_argument("--database-url", help="Database URL (optional)")
    args = parser.parse_args()
    return ServerConfig(args.host, args.port, args.database_url)


def run():
    """Run the server."""
    logger.info("Starting JFind service")
    config = parse_args()
    if config.database_url:
        os.environ["DATABASE_URL"] = config.database_url
    uvicorn.run(app, host=config.host, port=config.port)


if __name__ == "__main__":
    run()
