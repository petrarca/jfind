# JFind Service

A FastAPI-based service for managing and querying Java runtime environment scan results. The service provides REST API endpoints to store and retrieve information about Java installations across different computers.

This service works in conjunction with the [jfind scanner](./scanner/README.md), a cross-platform command-line tool that discovers and evaluates Java installations on systems. The scanner can automatically submit its findings to this service using its `-post` option.

## Features

- Store Java runtime scan results from multiple computers (submitted by the jfind scanner)
- Query scan results by computer name or scan ID
- Get latest scan results
- Retrieve Oracle Java runtime information
- OpenAPI documentation available at `/docs`

## Architecture

The JFind system consists of two main components:

1. **JFind Scanner**: A Go-based CLI tool that discovers Java installations on systems. See the [scanner documentation](./scanner/README.md) for details on:
   - Finding Java executables recursively
   - Evaluating Java version information
   - Detecting Oracle JDKs
   - Submitting results to this service

2. **JFind Service** (this component): A Python-based REST API service that:
   - Receives and stores scan results from the scanner
   - Provides query endpoints for analyzing Java installations across systems
   - Offers specialized queries for Oracle JDK installations

## Getting Started

### Prerequisites

- Python 3.x
- Virtual environment (recommended)

### Installation

1. Create and activate virtual environment:
```bash
python -m venv .venv
source .venv/bin/activate  # On Unix/macOS
```

2. Install dependencies:
```bash
task install:all
```

### Running the Service

You can start the service in several ways:

Using the task command:
```bash
task svc:run
```

Using the installed script (after installation):
```bash
jfind-svc [--host HOST] [--port PORT]
```

Or directly with Python:
```bash
python -m src.jfind_svc.main [--host HOST] [--port PORT]
```

Parameters:
- `--host`: Host address to bind to (default: "0.0.0.0")
- `--port`: Port number to listen on (default: 8000)

## API Endpoints

- `POST /jfind`: Submit Java runtime scan results
- `GET /jfind/scans`: Get latest scan results
- `GET /jfind/computer/{computer_name}`: Get scan results for a specific computer
- `GET /jfind/jdk/oracle`: Get all Oracle Java runtime information
- `GET /health`: Health check endpoint

For detailed API documentation, visit `http://localhost:8000/docs` after starting the service.