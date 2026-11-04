-- +goose Up
-- BELIRSIZ: offering price persistence is not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_add_offering_price_v1(
    p_authority_id uuid,
    p_version_id uuid,
    p_price_id uuid,
    p_cadence text,
    p_minor_units bigint,
    p_currency text,
    p_currency_registry_version bigint,
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

REVOKE ALL ON FUNCTION platform.billing_add_offering_price_v1(
    uuid, uuid, uuid, text, bigint, text,
    bigint, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_add_offering_price_v1(
    uuid, uuid, uuid, text, bigint, text,
    bigint, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_add_offering_price_v1(
    uuid, uuid, uuid, text, bigint, text,
    bigint, timestamptz
);
