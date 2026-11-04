-- +goose Up
-- BELIRSIZ: revision, version, lifecycle, and replay are not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_create_discount_v2(
    p_authority_id uuid,
    p_operation_id uuid,
    p_discount_id uuid,
    p_version_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_code text,
    p_name text,
    p_description text,
    p_available boolean,
    p_kind text,
    p_fixed_minor_units bigint,
    p_currency text,
    p_currency_registry_version bigint,
    p_basis_points bigint,
    p_duration text,
    p_repeating_cycles bigint,
    p_priority integer,
    p_global_use_limit bigint,
    p_per_tenant_use_limit bigint,
    p_all_offerings boolean,
    p_offering_count smallint,
    p_valid_start timestamptz,
    p_valid_end timestamptz,
    p_occurred_at timestamptz
)
RETURNS TABLE (
    discount_id uuid,
    version_id uuid,
    revision bigint,
    version bigint,
    lifecycle text,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        p_discount_id,
        p_version_id,
        0::bigint,
        0::bigint,
        ''::text,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_create_discount_v2(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, text, boolean, text,
    bigint, text, bigint, bigint, text, bigint,
    integer, bigint, bigint, boolean, smallint, timestamptz,
    timestamptz, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_discount_v2(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, text, boolean, text,
    bigint, text, bigint, bigint, text, bigint,
    integer, bigint, bigint, boolean, smallint, timestamptz,
    timestamptz, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_create_discount_v2(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, text, boolean, text,
    bigint, text, bigint, bigint, text, bigint,
    integer, bigint, bigint, boolean, smallint, timestamptz,
    timestamptz, timestamptz
);
