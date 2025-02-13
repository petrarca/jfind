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
- `--database-url`: Database URL to connect to (overrides environment variable)

### Database Configuration

The service supports both SQLite and PostgreSQL databases. The database connection can be configured in several ways, listed in order of priority:

1. Command line argument:
   ```bash
   jfind-svc --database-url "postgresql+asyncpg://user:pass@localhost:5432/jfind"
   ```

2. Environment variable (from shell or .env files):
   ```bash
   # Option 1: Set in shell
   export DATABASE_URL="postgresql+asyncpg://user:pass@localhost:5432/jfind"
   jfind-svc

   # Option 2: Set in .env file (in project directory)
   echo "DATABASE_URL=postgresql+asyncpg://user:pass@localhost:5432/jfind" > .env

   # Option 3: Set in ~/.env (user's home directory)
   echo "DATABASE_URL=postgresql+asyncpg://user:pass@localhost:5432/jfind" > ~/.env
   ```

3. Default configuration:
   - Development: SQLite (`sqlite+aiosqlite:///./jfind.db`)
   - Production: PostgreSQL (`postgresql+asyncpg://postgres:postgres@localhost:5432/jfind`)
   
   The environment is determined by the `ENV` environment variable (default: "development")

### Environment Files

The service supports loading environment variables from two locations:
1. `.env` in the project directory
2. `.env` in the user's home directory (`~/.env`)

These files can contain any environment variables used by the service, such as:
```env
DATABASE_URL=postgresql+asyncpg://user:pass@localhost:5432/jfind
ENV=production
```

## API Endpoints

- `POST /jfind`: Submit Java runtime scan results
- `GET /jfind/scans`: Get latest scan results
- `GET /jfind/computer/{computer_name}`: Get scan results for a specific computer
- `GET /jfind/oracle`: Get all Oracle Java runtime information
- `GET /jfind/oracle/{computer_name}`: Check if a specific computer has Oracle JDK installed
  - Response: 
    ```json
    {
      "computer_name": "string",
      "has_oracle": "true"|"false"|"unknown"
    }
    ```
  - "true": Computer has Oracle JDK installed
  - "false": Computer has Java records but no Oracle JDK
  - "unknown": No records found for this computer
- `GET /health`: Health check endpoint

### API Response Format

The scan results endpoints return data in the following format:

```json
{
  "meta": {
    "scan_ts": "2025-02-04T15:12:01Z",      // Scan timestamp in UTC
    "computer_name": "hostname",             // Name of the computer
    "user_name": "username",                 // Name of the user
    "scan_duration": "PT2.345S",            // Duration in ISO8601 format
    "has_oracle_jdk": false,                // Whether Oracle JDK was found
    "count_result": 2,                      // Number of Java installations found
    "count_require_license": 1,             // Number of Java installations requiring license
    "scanned_dirs": 56,                     // Number of directories scanned
    "scan_path": "/usr/lib/jvm"             // Starting path of the scan
  },
  "result": [
    {
      "java_executable": "/path/to/java",    // Path to Java executable
      "java_version": "11.0.20",            // Full Java version string (if -eval used)
      "java_vendor": "Oracle Corporation",   // Java vendor (if -eval used)
      "java_runtime": "Java(TM) SE Runtime", // Runtime name (if -eval used)
      "is_oracle": true,                     // Whether it's Oracle Java
      "java_version_major": 11,              // Major version number (8 for 1.8.0, 11 for 11.0.20)
      "java_version_update": 20,             // Update version number (202 for 1.8.0_202, 20 for 11.0.20)
      "exec_failed": true,                   // Present and true if java -version execution failed
      "require_license": true                // Present if license requirement is determined (true/false)
    }
  ]
}
```

For detailed API documentation, visit `http://localhost:8000/docs` after starting the service.