# JFind Service

A FastAPI-based service for managing and querying Java runtime environment scan results. The service provides REST API endpoints to store and retrieve information about Java installations across different computers.

## Features

- Store Java runtime scan results from multiple computers
- Query scan results by computer name or scan ID
- Get latest scan results
- Retrieve Oracle Java runtime information
- OpenAPI documentation available at `/docs`

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