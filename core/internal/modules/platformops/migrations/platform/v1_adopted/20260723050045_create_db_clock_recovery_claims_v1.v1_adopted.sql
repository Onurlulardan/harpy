-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_claim_recovery_manifest_v2(
    p_recovery_set_id uuid,
    p_tenant_id uuid,
    p_actor_user_id uuid,
    p_session_id uuid
) RETURNS TABLE(artifact_sha256 text, artifact_bytes bytea)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_runtime'
       OR NOT platform.iam_platform_current_owner_v1(
           p_tenant_id,p_actor_user_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery manifest claim denied';
    END IF;
    INSERT INTO platform.platformops_recovery_artifact_claims(
        artifact_kind,artifact_id,tenant_id,actor_user_id,session_id,
        artifact_sha256,claimed_at
    )
    SELECT 'recovery_manifest',recovery.recovery_set_id,p_tenant_id,
           p_actor_user_id,p_session_id,recovery.manifest_sha256,v_now
    FROM platform.platformops_recovery_sets AS recovery
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id=recovery.environment_id
     AND binding.tenant_id=p_tenant_id
    WHERE recovery.recovery_set_id=p_recovery_set_id
      AND recovery.state='protected' AND recovery.manifest_bytes IS NOT NULL
    ON CONFLICT DO NOTHING;
    IF NOT FOUND THEN RETURN; END IF;
    RETURN QUERY
    SELECT recovery.manifest_sha256,recovery.manifest_bytes
    FROM platform.platformops_recovery_sets AS recovery
    WHERE recovery.recovery_set_id=p_recovery_set_id;
END;
$function$;

CREATE FUNCTION platform.platformops_claim_restore_plan_v2(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_actor_user_id uuid,
    p_session_id uuid
) RETURNS TABLE(artifact_sha256 text, artifact_bytes bytea)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_runtime'
       OR NOT platform.iam_platform_current_owner_v1(
           p_tenant_id,p_actor_user_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore plan claim denied';
    END IF;
    INSERT INTO platform.platformops_recovery_artifact_claims(
        artifact_kind,artifact_id,tenant_id,actor_user_id,session_id,
        artifact_sha256,claimed_at
    )
    SELECT 'restore_plan',restore.operation_id,p_tenant_id,
           p_actor_user_id,p_session_id,restore.plan_sha256,v_now
    FROM platform.platformops_restore_operations AS restore
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id=restore.environment_id
     AND binding.tenant_id=p_tenant_id
    WHERE restore.operation_id=p_operation_id AND restore.state='plan_ready'
      AND restore.requested_by_actor_id=p_actor_user_id
      AND restore.requested_by_session_id=p_session_id
      AND restore.created_at+interval '15 minutes'>v_now
    ON CONFLICT DO NOTHING;
    IF NOT FOUND THEN RETURN; END IF;
    RETURN QUERY
    SELECT restore.plan_sha256,restore.plan_bytes
    FROM platform.platformops_restore_operations AS restore
    WHERE restore.operation_id=p_operation_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_claim_recovery_manifest_v2(
    uuid,uuid,uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_claim_restore_plan_v2(
    uuid,uuid,uuid,uuid
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_claim_recovery_manifest_v1(
    uuid,uuid,uuid,uuid,timestamptz
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.platformops_claim_restore_plan_v1(
    uuid,uuid,uuid,uuid,timestamptz
) FROM harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_recovery_manifest_v2(
    uuid,uuid,uuid,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_restore_plan_v2(
    uuid,uuid,uuid,uuid
) TO harpy_runtime;
