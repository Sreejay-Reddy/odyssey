from .results import BuildLedgerResult
from psycopg.types.json import Jsonb

class BuildLedger:
    def __init__(self, get_conn, key, steps):
        self.get_conn = get_conn
        self.key = key
        self.steps = steps

    async def run(self):

        ledger_rows = []
        delivery_rows = []

        for sequence, step in enumerate(self.steps, start=1):

            mode = "delegated" if step.delegate else "local"

            ledger_rows.append(
                (
                    self.key,
                    step.target,
                    sequence,
                    mode,
                    Jsonb(step.kwargs),
                )
            )

            if step.delegate:
                delivery_rows.append(
                    (
                        self.key,
                        step.target,
                        step.delegate,
                    )
                )

        async with self.get_conn() as conn:
            try:
                async with conn.cursor() as cur:

                    # Parent ledger
                    await cur.execute(
                        """
                        INSERT INTO odyssey_ledger(key)
                        VALUES(%s)
                        """,
                        (self.key,),
                    )

                    # Individual journeys
                    await cur.executemany(
                        """
                        INSERT INTO odyssey_journeys(
                            key,
                            target,
                            sequence,
                            mode,
                            input
                        )
                        VALUES(%s, %s, %s, %s, %s)
                        """,
                        ledger_rows,
                    )

                    # Delegated deliveries
                    if delivery_rows:
                        await cur.executemany(
                            """
                            INSERT INTO odyssey_deliveries(
                                key,
                                target,
                                emit_to
                            )
                            VALUES(%s, %s, %s)
                            """,
                            delivery_rows,
                        )

                await conn.commit()

                return BuildLedgerResult(
                    key=self.key,
                    targets=[
                        step.target
                        for step in self.steps
                    ],
                    delegated=[
                        step.target
                        for step in self.steps
                        if step.delegate
                    ],
                    local=[
                        step.target
                        for step in self.steps
                        if not step.delegate
                    ],
                )

            except Exception as e:
                await conn.rollback()
                raise RuntimeError("Failed to build ledger") from e