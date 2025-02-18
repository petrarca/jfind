"""Models for the JFind service."""

from pydantic import BaseModel


class JavaRuntime(BaseModel):
    """Model for Java runtime information."""

    java_executable: str
    java_runtime: str | None = None
    java_vendor: str | None = None
    is_oracle: bool | None = None
    java_version: str | None = None
    java_version_major: int | None = None
    java_version_update: int | None = None
    require_license: bool | None = None


class MetaInfo(BaseModel):
    """Model for scan metadata."""

    scan_ts: str
    computer_name: str
    user_name: str
    scan_duration: str
    has_oracle_jdk: bool
    count_result: int
    count_require_license: int
    scanned_dirs: int
    scan_path: str
    platform_info: str


class ScannerResults(BaseModel):
    """Model for scanner results matching the Go scanner's JSONOutput structure."""

    meta: MetaInfo
    result: list[JavaRuntime]
