-- +goose Up
-- BELIRSIZ: legacy identity aliases are not determined by the caller.
CREATE FUNCTION platform.billing_legacy_identity_alias_v1()
RETURNS TABLE (
    legacy_id uuid,
    canonical_id uuid,
    record_kind text,
    authority_id uuid,
    scope_kind text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::uuid,
        NULL::uuid,
        NULL::text,
        NULL::uuid,
        NULL::text
    WHERE false;
$$;

REVOKE ALL ON FUNCTION platform.billing_legacy_identity_alias_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_legacy_identity_alias_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_legacy_identity_alias_v1();
