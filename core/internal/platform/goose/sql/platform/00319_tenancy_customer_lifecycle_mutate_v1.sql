-- +goose Up
CREATE FUNCTION platform.tenancy_customer_lifecycle_mutate_v1(
    in_environment_id uuid,
    in_control_tenant_id uuid,
    in_actor_id uuid,
    in_target_tenant_id uuid,
    in_kind text,
    in_name text,
    in_expected_revision bigint,
    in_idempotency_key text,
    in_request_sha256 bytea,
    in_now timestamptz
)
RETURNS TABLE (
    outcome text,
    replay boolean,
    id uuid,
    name text,
    slug text,
    status text,
    lifecycle_revision bigint,
    settings_revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    offboarded_at timestamptz,
    purged_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        false,
        NULL::uuid,
        in_name,
        NULL::text,
        NULL::text,
        NULL::bigint,
        NULL::bigint,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz
    ;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_mutate_v1(uuid, uuid, uuid, uuid, text, text, bigint, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_mutate_v1(uuid, uuid, uuid, uuid, text, text, bigint, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_lifecycle_mutate_v1(uuid, uuid, uuid, uuid, text, text, bigint, text, bytea, timestamptz);
