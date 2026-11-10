-- +goose Up
CREATE TABLE platform.environment_identity (
    singleton boolean PRIMARY KEY,
    environment_id uuid NOT NULL,
    database_origin text NOT NULL,
    database_system_identifier text NOT NULL,
    database_oid bigint NOT NULL,
    release_id uuid NOT NULL,
    CHECK (singleton)
);

CREATE TABLE platform.extension_registry (
    name text PRIMARY KEY,
    version text NOT NULL,
    schema_name text NOT NULL,
    owner_role text NOT NULL,
    database_origin text NOT NULL,
    verified_run_id uuid NOT NULL
);

REVOKE ALL ON TABLE
    platform.environment_identity,
    platform.extension_registry
FROM PUBLIC;
GRANT SELECT ON TABLE
    platform.environment_identity,
    platform.extension_registry
TO harpy_api, harpy_worker;

INSERT INTO platform.environment_identity (
    singleton, environment_id, database_origin,
    database_system_identifier, database_oid, release_id
)
SELECT
    TRUE,
    COALESCE(
        (SELECT policy.environment_id FROM platform.tenancy_deployment_policy AS policy LIMIT 1),
        pg_catalog.gen_random_uuid()
    ),
    'baseline',
    control.system_identifier::text,
    database_row.oid::bigint,
    pg_catalog.gen_random_uuid()
FROM pg_catalog.pg_control_system() AS control
CROSS JOIN pg_catalog.pg_database AS database_row
WHERE database_row.datname = pg_catalog.current_database();

INSERT INTO platform.extension_registry (
    name, version, schema_name, owner_role, database_origin, verified_run_id
)
SELECT
    extension.extname,
    extension.extversion,
    namespace.nspname,
    pg_catalog.pg_get_userbyid(extension.extowner),
    identity.database_origin,
    identity.release_id
FROM pg_catalog.pg_extension AS extension
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = extension.extnamespace
CROSS JOIN platform.environment_identity AS identity
WHERE extension.extname = 'vector';

-- +goose Down
DROP TABLE platform.extension_registry;
DROP TABLE platform.environment_identity;
