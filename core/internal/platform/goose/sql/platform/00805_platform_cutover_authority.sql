-- +goose Up
CREATE TABLE platform.cutover_groups (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    generation bigint,
    state text,
    writer text,
    requested_by uuid,
    approved_by uuid,
    reason text,
    change_reference text,
    acceptance_digest text,
    changed_at timestamptz,
    PRIMARY KEY (scope_kind, scope_id, group_id)
);

CREATE TABLE platform.domain_cutovers (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    domain text NOT NULL,
    generation bigint,
    state text,
    writer text,
    requested_by uuid,
    approved_by uuid,
    reason text,
    change_reference text,
    acceptance_digest text,
    changed_at timestamptz,
    PRIMARY KEY (scope_kind, scope_id, domain)
);

CREATE TABLE platform.cutover_group_members (
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    domain text NOT NULL,
    ordinal bigint,
    PRIMARY KEY (scope_kind, scope_id, group_id, domain)
);

CREATE TABLE platform.cutover_job_bindings (
    job_id uuid PRIMARY KEY,
    tenant_id uuid,
    group_generation bigint,
    cutover_generation bigint,
    payload_ref text,
    payload_sha256 text,
    operation_id uuid,
    cutover_run_id uuid,
    lease_manifest_sha256 text
);

REVOKE ALL ON TABLE
    platform.cutover_groups,
    platform.cutover_group_members,
    platform.domain_cutovers,
    platform.cutover_job_bindings
FROM PUBLIC;
GRANT SELECT ON TABLE
    platform.cutover_groups,
    platform.cutover_group_members,
    platform.domain_cutovers,
    platform.cutover_job_bindings
TO harpy_api, harpy_worker;

-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_cutover_shadow_operation_active_v1(
    in_operation_id uuid,
    in_cutover_run_id uuid,
    in_scope_kind text,
    in_scope_id uuid,
    in_group_id text,
    in_domain text,
    in_cutover_generation bigint,
    in_lease_manifest_sha256 text,
    in_observed_at timestamptz
)
RETURNS boolean
LANGUAGE sql
STABLE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_cutover_shadow_operation_active_v1(
    uuid, uuid, text, uuid, text, text, bigint, text, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_cutover_shadow_operation_active_v1(
    uuid, uuid, text, uuid, text, text, bigint, text, timestamptz
) TO harpy_api, harpy_worker;

ALTER TABLE platform.jobs
    ADD COLUMN created_at timestamptz DEFAULT pg_catalog.clock_timestamp();
UPDATE platform.jobs
SET created_at = COALESCE(updated_at, pg_catalog.clock_timestamp())
WHERE created_at IS NULL;

ALTER TABLE platform.outbox
    ADD COLUMN parked_at timestamptz,
    ADD COLUMN created_at timestamptz DEFAULT pg_catalog.clock_timestamp();
UPDATE platform.outbox
SET created_at = COALESCE(occurred_at, pg_catalog.clock_timestamp())
WHERE created_at IS NULL;

-- +goose Down
ALTER TABLE platform.outbox DROP COLUMN created_at, DROP COLUMN parked_at;
ALTER TABLE platform.jobs DROP COLUMN created_at;
DROP FUNCTION platform.platform_async_cutover_shadow_operation_active_v1(
    uuid, uuid, text, uuid, text, text, bigint, text, timestamptz
);
DROP TABLE platform.cutover_job_bindings;
DROP TABLE platform.cutover_group_members;
DROP TABLE platform.domain_cutovers;
DROP TABLE platform.cutover_groups;
