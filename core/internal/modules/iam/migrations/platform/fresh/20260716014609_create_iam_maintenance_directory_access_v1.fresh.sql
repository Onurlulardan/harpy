-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_maintenance_directory_purge_v1(
    p_tenant_id uuid, p_identifier_id uuid, p_source_revision bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_rows bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL OR p_identifier_id IS NULL OR
       p_source_revision IS NULL OR p_source_revision < 1 OR
       NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id = p_tenant_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM maintenance directory purge input is invalid';
    END IF;
    DELETE FROM platform.iam_login_directory AS directory
    WHERE directory.tenant_id = p_tenant_id
      AND directory.source_identifier_id = p_identifier_id
      AND directory.state = 'tombstoned'
      AND directory.tombstoned_at <=
          pg_catalog.transaction_timestamp() - pg_catalog.make_interval(days => 30)
      AND directory.source_revision <= p_source_revision;
    GET DIAGNOSTICS v_rows = ROW_COUNT;
    RETURN v_rows;
END;
$function$;

REVOKE ALL ON FUNCTION platform.iam_maintenance_directory_purge_v1(
    uuid,uuid,bigint
) FROM PUBLIC;

CREATE FUNCTION platform.iam_maintenance_directory_snapshot_v1(p_tenant_id uuid)
RETURNS TABLE (
    source_identifier_id uuid, lookup_key_id text, identifier_hmac bytea,
    user_id uuid, source_revision bigint, identifier_kind text,
    state text, source_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL OR NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id = p_tenant_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM maintenance directory snapshot input is invalid';
    END IF;
    RETURN QUERY
    SELECT directory.source_identifier_id, directory.lookup_key_id,
           directory.identifier_hmac, directory.user_id,
           directory.source_revision, directory.identifier_kind,
           directory.state, directory.source_sha256
    FROM platform.iam_login_directory AS directory
    WHERE directory.tenant_id = p_tenant_id
    ORDER BY directory.source_identifier_id, directory.lookup_key_id;
END;
$function$;

REVOKE ALL ON FUNCTION platform.iam_maintenance_directory_snapshot_v1(uuid)
FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_maintenance_directory_apply_v1(
    uuid,uuid,uuid,text,text,bigint,bytea,jsonb
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_maintenance_directory_purge_v1(
    uuid,uuid,bigint
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.iam_maintenance_directory_snapshot_v1(uuid)
TO forgecrud_runtime;
-- +goose StatementEnd
