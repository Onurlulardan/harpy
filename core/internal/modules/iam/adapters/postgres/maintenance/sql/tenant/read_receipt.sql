WITH receipt AS (
    SELECT * FROM iam_maintenance_receipt_read_v1(
        pg_catalog.current_schema(), $1, $2, $3, $4, $5, $6
    )
)
SELECT true, receipt.* FROM receipt
UNION ALL
SELECT false, NULL::uuid, NULL::timestamptz, NULL::text,
       NULL::bytea, NULL::text, NULL::integer
WHERE NOT EXISTS (SELECT 1 FROM receipt)
