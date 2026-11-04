-- +goose Up
-- BELIRSIZ: revision, version, lifecycle, and replay are not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_create_offering_v3(
    p_authority_id uuid,
    p_operation_id uuid,
    p_offering_id uuid,
    p_version_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 bytea,
    p_offering_key text,
    p_name text,
    p_description text,
    p_sort_order integer,
    p_highlight_label text,
    p_color text,
    p_billing_interval text,
    p_public boolean,
    p_available boolean,
    p_base boolean,
    p_trial_days integer,
    p_grace_days integer,
    p_price_count smallint,
    p_grant_count smallint,
    p_effective_at timestamptz,
    p_effective_end timestamptz,
    p_occurred_at timestamptz
)
RETURNS TABLE (
    offering_id uuid,
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
        p_offering_id,
        p_version_id,
        0::bigint,
        0::bigint,
        ''::text,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_create_offering_v3(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, text, integer, text,
    text, text, boolean, boolean, boolean, integer,
    integer, smallint, smallint, timestamptz, timestamptz, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_offering_v3(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, text, integer, text,
    text, text, boolean, boolean, boolean, integer,
    integer, smallint, smallint, timestamptz, timestamptz, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_create_offering_v3(
    uuid, uuid, uuid, uuid, uuid, text,
    bytea, text, text, text, integer, text,
    text, text, boolean, boolean, boolean, integer,
    integer, smallint, smallint, timestamptz, timestamptz, timestamptz
);
