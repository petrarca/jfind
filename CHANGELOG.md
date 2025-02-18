# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## General
To upgrade existing database:
```bash
# Apply the new migration
alembic upgrade head
```

## [Unreleased]

### Added
- Platform information retrieval by scanner
  - Added `platform_info` field to scan results
  - Captures OS details, version, architecture, and system name
  - Example: `OS=linux;Version=5.15.0;Arch=amd64;Name=Ubuntu 22.04.2 LTS`

### Changed
- Database model and API updates
  - Added `most_recent` flag to `ScanInfo` table
  - Only one scan per computer can have `most_recent=True`
  - New scans automatically set previous scans' `most_recent` to `False`

- API behavior changes
  - `/jfind/scans` now returns only most recent scan per computer
  - `/jfind/scans/{computer_name}` defaults to most recent scan only
    - Added `limit` parameter to control number of results:
      - `limit=0` (default): Return only most recent scan
      - `limit=-1`: Return all scans
      - `limit=N`: Return N most recent scans
  - `/jfind/oracle` endpoint now only returns data from most recent scans
  - `/jfind/require_license/{computer_name}` checks based on most recent scan only
