"""add scan_path and count_require_license columns to scan_info

Revision ID: 26be0b9c206b
Revises: cb1040c5742e
Create Date: 2025-02-14 10:38:50.532733

"""

from typing import Sequence, Union

import sqlalchemy as sa

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "26be0b9c206b"
down_revision: Union[str, None] = "cb1040c5742e"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Add scan_path column
    op.add_column("scan_info", sa.Column("scan_path", sa.String(length=1024), nullable=True))
    # Add count_require_license column
    op.add_column("scan_info", sa.Column("count_require_license", sa.Integer(), nullable=True))


def downgrade() -> None:
    # Remove the columns
    op.drop_column("scan_info", "count_require_license")
    op.drop_column("scan_info", "scan_path")
