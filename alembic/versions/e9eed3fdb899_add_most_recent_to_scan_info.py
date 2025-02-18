"""add most_recent to scan_info

Revision ID: e9eed3fdb899
Revises: b85d9e8ebb87
Create Date: 2025-02-18 18:39:46

"""

from typing import Sequence, Union

import sqlalchemy as sa
from sqlalchemy import alias, column, select, table

from alembic import op

# revision identifiers, used by Alembic.
revision: str = "e9eed3fdb899"
down_revision: Union[str, None] = "b85d9e8ebb87"
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def update_most_recent(session: sa.engine.Connection) -> None:
    """Update most_recent column based on latest scan_ts per computer"""
    # Define table for the update
    scan_info = table(
        "scan_info",
        column("id", sa.Integer),
        column("computer_name", sa.String),
        column("scan_ts", sa.DateTime),
        column("most_recent", sa.Boolean),
    )

    # Create an alias for the self-join
    si2 = alias(scan_info)

    # Update statement using a correlated subquery
    # For each row, check if there exists a newer scan for the same computer
    subq = (
        select(1).where(sa.and_(si2.c.computer_name == scan_info.c.computer_name, si2.c.scan_ts > scan_info.c.scan_ts)).exists()
    )

    update_stmt = scan_info.update().values(most_recent=~subq)

    # Execute the update
    session.execute(update_stmt)


def upgrade() -> None:
    """Add most_recent column with default False"""
    # Add most_recent column with default False
    op.add_column("scan_info", sa.Column("most_recent", sa.Boolean(), server_default=sa.false(), nullable=False))

    # Update most_recent column based on latest scan_ts per computer
    update_most_recent(op.get_bind())


def downgrade() -> None:
    op.drop_column("scan_info", "most_recent")
