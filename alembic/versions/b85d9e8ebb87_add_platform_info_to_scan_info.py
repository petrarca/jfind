"""add platform_info to scan_info

Revision ID: b85d9e8ebb87
Revises: 26be0b9c206b
Create Date: 2025-02-18 14:28:24.000000

"""

from typing import Sequence, Union

import sqlalchemy as sa

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "b85d9e8ebb87"
down_revision: Union[str, None] = "26be0b9c206b"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Add platform_info column to scan_info table
    op.add_column("scan_info", sa.Column("platform_info", sa.String(length=255), nullable=True))


def downgrade() -> None:
    # Remove platform_info column from scan_info table
    op.drop_column("scan_info", "platform_info")
