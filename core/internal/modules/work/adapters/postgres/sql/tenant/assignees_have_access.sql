WITH supplied AS (
    SELECT principal.kind, principal.id
    FROM ROWS FROM (
        pg_catalog.unnest($2::text[]),
        pg_catalog.unnest($3::uuid[])
    ) AS principal(kind, id)
), locked_memberships AS MATERIALIZED (
    SELECT membership.principal_kind, membership.principal_id
    FROM work_board_memberships AS membership
    WHERE membership.board_id = $1
      AND membership.active
    FOR SHARE
)
SELECT NOT EXISTS (
    SELECT 1
    FROM supplied AS principal
    LEFT JOIN locked_memberships AS membership
      ON membership.principal_kind = principal.kind
     AND membership.principal_id = principal.id
    WHERE membership.principal_id IS NULL
)
