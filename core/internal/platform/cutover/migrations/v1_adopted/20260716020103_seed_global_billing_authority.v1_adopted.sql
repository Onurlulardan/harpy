-- +goose Up
WITH identity AS (
    SELECT environment_id,
           (CASE WHEN pg_catalog.left(environment_id::text, 1) = '0'
               THEN '1' ELSE '0' END ||
            pg_catalog.substr(environment_id::text, 2))::uuid AS approved_by
    FROM platform.environment_identity
    WHERE singleton = TRUE AND database_origin = 'fresh'
)
INSERT INTO platform.domain_cutovers (
    scope_kind, scope_id, domain, state, writer, generation,
    requested_by, approved_by, reason, change_reference, acceptance_digest
)
SELECT 'global', environment_id, 'billing', 'v2', 'v2', 1,
       environment_id, approved_by, 'fresh V2 environment bootstrap',
       'migration:20260716020103:fresh',
       'fbeb43a0f1716baa3566423ad6b3f441e305bc6b6f8cf73c4ddfdf8535c80694'
FROM identity;

WITH identity AS (
    SELECT environment_id,
           (CASE WHEN pg_catalog.left(environment_id::text, 1) = '0'
               THEN '1' ELSE '0' END ||
            pg_catalog.substr(environment_id::text, 2))::uuid AS approved_by
    FROM platform.environment_identity
    WHERE singleton = TRUE AND database_origin = 'fresh'
)
INSERT INTO platform.cutover_groups (
    scope_kind, scope_id, group_id, state, writer, generation,
    requested_by, approved_by, reason, change_reference, acceptance_digest
)
SELECT 'global', environment_id, 'billing', 'v2', 'v2', 1,
       environment_id, approved_by, 'fresh V2 environment bootstrap',
       'migration:20260716020103:fresh',
       'fbeb43a0f1716baa3566423ad6b3f441e305bc6b6f8cf73c4ddfdf8535c80694'
FROM identity;

INSERT INTO platform.cutover_group_members (
    scope_kind, scope_id, group_id, ordinal, domain
)
SELECT 'global', environment_id, 'billing', 1, 'billing'
FROM platform.environment_identity
WHERE singleton = TRUE AND database_origin = 'fresh';

SELECT 1 / ((
    (SELECT pg_catalog.count(*) FROM platform.domain_cutovers
     WHERE scope_kind='global' AND domain='billing' AND state='v2' AND writer='v2')=1
    AND (SELECT pg_catalog.count(*) FROM platform.cutover_groups
         WHERE scope_kind='global' AND group_id='billing'
           AND state='v2' AND writer='v2')=1
    AND (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
         WHERE scope_kind='global' AND group_id='billing'
           AND ordinal=1 AND domain='billing')=1
    )::pg_catalog.int4) AS fresh_global_billing_authority_cardinality;
