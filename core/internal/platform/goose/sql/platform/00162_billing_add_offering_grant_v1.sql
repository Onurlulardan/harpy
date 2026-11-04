-- +goose Up
-- BELIRSIZ: offering grant persistence is not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_add_offering_grant_v1(
    p_authority_id uuid,
    p_version_id uuid,
    p_entitlement_key text,
    p_value_kind text,
    p_boolean_value boolean,
    p_integer_value bigint,
    p_unlimited boolean,
    p_occurred_at timestamptz
)
RETURNS void
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT CAST(NULL AS pg_catalog.void);
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_add_offering_grant_v1(
    uuid, uuid, text, text, boolean, bigint,
    boolean, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_add_offering_grant_v1(
    uuid, uuid, text, text, boolean, bigint,
    boolean, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_add_offering_grant_v1(
    uuid, uuid, text, text, boolean, bigint,
    boolean, timestamptz
);
