-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_worker_capability_consume_v1(
    p_plan_id uuid, p_resource_kind text, p_resource_id uuid,
    p_actual_source_schema_id uuid, p_source_revision bigint,
    p_source_sha256 bytea, p_target_schema_id uuid, p_target_sha256 bytea
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_consumed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_plan_id IS NULL OR
       p_resource_kind NOT IN ('user','organization','role') OR
       p_resource_id IS NULL OR p_actual_source_schema_id IS NULL OR
       p_source_revision < 1 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 OR
       p_target_schema_id IS NULL OR
       pg_catalog.octet_length(p_target_sha256) IS DISTINCT FROM 32 THEN
        RETURN false;
    END IF;
    DELETE FROM iam_profile_backfill_worker_capabilities AS capability
    WHERE capability.backend_pid = pg_catalog.pg_backend_pid()
      AND capability.transaction_id = pg_catalog.pg_current_xact_id()
      AND capability.plan_id = p_plan_id
      AND capability.resource_kind = p_resource_kind
      AND capability.resource_id = p_resource_id
      AND capability.actual_source_schema_id = p_actual_source_schema_id
      AND capability.source_revision = p_source_revision
      AND capability.source_sha256 = p_source_sha256
      AND capability.target_schema_id = p_target_schema_id
      AND capability.target_sha256 = p_target_sha256
    RETURNING true INTO v_consumed;
    RETURN coalesce(v_consumed, false);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_worker_capability_consume_v1(
    uuid,text,uuid,uuid,bigint,bytea,uuid,bytea
) FROM PUBLIC;
