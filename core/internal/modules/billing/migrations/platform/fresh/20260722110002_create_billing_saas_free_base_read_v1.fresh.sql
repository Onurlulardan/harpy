-- +goose Up
ALTER TABLE platform.billing_offering_entitlements ENABLE ALWAYS TRIGGER billing_offering_entitlements_authority_guard;
ALTER TABLE platform.billing_price_versions ENABLE ALWAYS TRIGGER billing_price_versions_authority_guard;
ALTER TABLE platform.billing_offering_versions ENABLE ALWAYS TRIGGER billing_offering_versions_authority_guard;
ALTER TABLE platform.billing_offerings ENABLE ALWAYS TRIGGER billing_offerings_authority_guard;

-- +goose StatementBegin
CREATE FUNCTION platform.billing_saas_free_base_read_v1()
RETURNS TABLE (
    offering_version_id uuid, revision bigint, offering_sha256 text,
    effective_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='SaaS free base read denied';
    END IF;
    RETURN QUERY SELECT publication.offering_version_id,publication.revision,
        publication.offering_sha256,publication.effective_at
    FROM platform.billing_saas_free_base_publications AS publication
    WHERE publication.offering_key='saas_free_base_v1';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='SaaS free base publication is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_saas_free_base_read_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_saas_free_base_read_v1()
TO forgecrud_v2_runtime;
