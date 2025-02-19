"""Models for the JFind service."""

from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field, field_serializer, field_validator


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

    model_config = ConfigDict(from_attributes=True)


class MetaInfo(BaseModel):
    """Model for scan metadata."""

    scan_ts: datetime
    computer_name: str
    user_name: str
    scan_duration: str
    has_oracle_jdk: bool
    count_result: int
    count_require_license: int
    scanned_dirs: int
    scan_path: str
    platform_info: str | None
    scan_id: int | None = Field(alias="id", default=None)

    model_config = ConfigDict(
        from_attributes=True,
        populate_by_name=True,
    )

    @field_serializer("scan_ts")
    def serialize_ts(self, ts: datetime | None, _info) -> str | None:
        return ts.isoformat() if ts is not None else None

    @field_validator("scan_ts", mode="before")
    @classmethod
    def parse_scan_ts(cls, value: str | datetime | None) -> datetime | None:
        return value if isinstance(value, datetime) else _isostr_to_datetime(value)


class ScannerResult(BaseModel):
    """Model for scanner results matching the Go scanner's JSONOutput structure."""

    meta: MetaInfo
    runtimes: list[JavaRuntime]

    model_config = ConfigDict(from_attributes=True)


def _isostr_to_datetime(value: str | None) -> datetime | None:
    if value is None:
        return None
    try:
        # Parse the ISO format string into a datetime object
        return datetime.fromisoformat(value)
    except (TypeError, ValueError) as e:
        raise ValueError(f"Invalid datetime format: {value}") from e
