WITH target_relation AS (
    SELECT relation.relacl, relation.relowner
    FROM pg_catalog.pg_class AS relation
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = relation.relnamespace
    WHERE namespace.nspname = $1
      AND relation.relname = $2
      AND relation.relkind = 'r'
), table_acl AS (
    SELECT COALESCE(grantee.rolname, 'PUBLIC') AS grantee,
           access.privilege_type,
           access.is_grantable
    FROM target_relation AS relation
    CROSS JOIN LATERAL pg_catalog.aclexplode(
        COALESCE(relation.relacl,
                 pg_catalog.acldefault('r', relation.relowner))
    ) AS access
    LEFT JOIN pg_catalog.pg_roles AS grantee
      ON grantee.oid = access.grantee
), target_namespace AS (
    SELECT namespace.nspacl, namespace.nspowner
    FROM pg_catalog.pg_namespace AS namespace
    WHERE namespace.nspname = $1
), schema_acl AS (
    SELECT COALESCE(grantee.rolname, 'PUBLIC') AS grantee,
           access.privilege_type,
           access.is_grantable
    FROM target_namespace AS namespace
    CROSS JOIN LATERAL pg_catalog.aclexplode(
        COALESCE(namespace.nspacl,
                 pg_catalog.acldefault('n', namespace.nspowner))
    ) AS access
    LEFT JOIN pg_catalog.pg_roles AS grantee
      ON grantee.oid = access.grantee
)
SELECT
    (SELECT pg_catalog.count(*) FROM table_acl
      WHERE grantee = 'harpy_forms_record_reader') = 1
    AND (SELECT pg_catalog.count(*) FROM table_acl
      WHERE grantee = 'harpy_forms_record_reader'
        AND privilege_type = 'SELECT' AND NOT is_grantable) = 1
    AND (SELECT pg_catalog.count(*) FROM table_acl
      WHERE grantee = 'harpy_forms_record_writer') = 4
    AND (SELECT pg_catalog.count(*) FROM table_acl
      WHERE grantee = 'harpy_forms_record_writer'
        AND privilege_type IN ('SELECT', 'INSERT', 'UPDATE', 'DELETE')
        AND NOT is_grantable) = 4
    AND (SELECT pg_catalog.count(*) FROM table_acl
      WHERE grantee IN ('PUBLIC', 'harpy_runtime', 'harpy_v2_runtime')) = 0,
    (SELECT pg_catalog.count(*) FROM schema_acl
      WHERE grantee = 'harpy_forms_record_reader') = 1
    AND (SELECT pg_catalog.count(*) FROM schema_acl
      WHERE grantee = 'harpy_forms_record_reader'
        AND privilege_type = 'USAGE' AND NOT is_grantable) = 1
    AND (SELECT pg_catalog.count(*) FROM schema_acl
      WHERE grantee = 'harpy_forms_record_writer') = 1
    AND (SELECT pg_catalog.count(*) FROM schema_acl
      WHERE grantee = 'harpy_forms_record_writer'
        AND privilege_type = 'USAGE' AND NOT is_grantable) = 1;
