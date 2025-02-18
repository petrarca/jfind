"""initial schema

Revision ID: cb1040c5742e
Revises:
Create Date: 2025-02-14 10:40:33.123456

"""

from typing import Sequence, Union

import sqlalchemy as sa

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "cb1040c5742e"
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create scan_info table
    op.create_table(
        "scan_info",
        sa.Column("id", sa.Integer(), nullable=False),
        sa.Column("scan_ts", sa.DateTime(), nullable=False),
        sa.Column("computer_name", sa.String(length=255), nullable=False),
        sa.Column("user_name", sa.String(length=255), nullable=False),
        sa.Column("scan_duration", sa.String(length=50), nullable=False),
        sa.Column("has_oracle_jdk", sa.Boolean(), nullable=False),
        sa.Column("count_result", sa.Integer(), nullable=False),
        sa.Column("scanned_dirs", sa.Integer(), nullable=False),
        sa.Column("created_at", sa.DateTime(), nullable=False),
        sa.PrimaryKeyConstraint("id"),
    )

    # Create java_info table
    op.create_table(
        "java_info",
        sa.Column("id", sa.Integer(), nullable=False),
        sa.Column("scan_id", sa.Integer(), nullable=False),
        sa.Column("computer_name", sa.String(length=255), nullable=False),
        sa.Column("java_executable", sa.String(length=1024), nullable=False),
        sa.Column("java_runtime", sa.String(length=255), nullable=True),
        sa.Column("java_vendor", sa.String(length=255), nullable=True),
        sa.Column("is_oracle", sa.Boolean(), nullable=True),
        sa.Column("java_version", sa.String(length=50), nullable=True),
        sa.Column("java_version_major", sa.Integer(), nullable=True),
        sa.Column("java_version_update", sa.Integer(), nullable=True),
        sa.Column("require_license", sa.Boolean(), nullable=True),
        sa.Column("created_at", sa.DateTime(), nullable=False),
        sa.ForeignKeyConstraint(
            ["scan_id"],
            ["scan_info.id"],
        ),
        sa.PrimaryKeyConstraint("id"),
    )


def downgrade() -> None:
    op.drop_table("java_info")
    op.drop_table("scan_info")
