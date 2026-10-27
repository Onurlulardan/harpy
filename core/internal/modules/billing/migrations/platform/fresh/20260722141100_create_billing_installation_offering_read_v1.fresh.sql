-- +goose Up
-- Migration stream version: 20260722141100.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_offering_read_v1(
    p_environment_id uuid, p_deployment_mode text
)
RETURNS TABLE (
    offering_version_id uuid, revision bigint, offering_sha256 text,
    effective_at timestamptz, source_sequence bigint, source_sha256 text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_count bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_runtime'
       OR p_environment_id IS NULL
       OR p_deployment_mode NOT IN ('onpremise','saas') THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='billing installation offering read denied';
    END IF;
    IF p_deployment_mode='onpremise' THEN
        RETURN QUERY
        SELECT publication.offering_version_id,publication.revision,
            publication.offering_sha256,publication.effective_at,
            publication.snapshot_sequence,publication.snapshot_sha256
        FROM platform.billing_license_heads AS head
        JOIN platform.billing_license_offering_publications AS publication
          ON publication.license_id=head.license_id
         AND publication.environment_id=head.environment_id
         AND publication.snapshot_sequence=head.sequence
         AND publication.snapshot_sha256=head.payload_sha256
        WHERE head.environment_id=p_environment_id;
    ELSE
        RETURN QUERY
        SELECT publication.offering_version_id,publication.revision,
            publication.offering_sha256,publication.effective_at,
            NULL::bigint,NULL::text
        FROM platform.billing_saas_control_offering_publications AS publication
        WHERE publication.offering_key='saas_platform_control_base_v1';
    END IF;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count IS DISTINCT FROM 1::bigint THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='billing installation offering is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.billing_installation_offering_read_v1(uuid,text)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_offering_read_v1(uuid,text)
TO forgecrud_runtime;
