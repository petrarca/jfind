"""JFind Service main module."""

import argparse
from contextlib import asynccontextmanager
from typing import NamedTuple

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from jfind_svc.db import init_db
from jfind_svc.routes import health as health_router
from jfind_svc.routes import router as api_router


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
app.include_router(api_router, prefix="/api")  # /api/* endpoints


class ServerConfig(NamedTuple):
    """Server configuration."""

    host: str = "0.0.0.0"
    port: int = 8000


def parse_args() -> ServerConfig:
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(description="JFind Service")
    parser.add_argument("--port", type=int, default=8000, help="Port to run the server on (default: 8000)")
    parser.add_argument("--host", type=str, default="0.0.0.0", help="Host to run the server on (default: 0.0.0.0)")

    # Don't exit on error, just use defaults
    args, _ = parser.parse_known_args()
    return ServerConfig(host=args.host, port=args.port)


def run():
    """Entry point for the application."""
    config = parse_args()
    uvicorn.run(app, host=config.host, port=config.port)


if __name__ == "__main__":
    run()
