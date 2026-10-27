-- +goose Up
-- +goose StatementBegin
ALTER TABLE platform.platformops_restore_operations
    ADD COLUMN IF NOT EXISTS requested_by_actor_id uuid,
    ADD COLUMN IF NOT EXISTS requested_by_session_id uuid;

CREATE TABLE IF NOT EXISTS platform.platformops_recovery_artifact_claims (
    artifact_kind text NOT NULL
        CHECK (artifact_kind IN ('recovery_manifest','restore_plan')),
    artifact_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    session_id uuid NOT NULL,
    artifact_sha256 text NOT NULL CHECK (artifact_sha256 ~ '^[0-9a-f]{64}$'),
    claimed_at timestamptz NOT NULL,
    PRIMARY KEY (artifact_kind, artifact_id)
);

CREATE FUNCTION platform.platformops_list_recovery_sets_v1(p_environment_id uuid)
RETURNS TABLE(
    recovery_set_id uuid, operation_id uuid, environment_id uuid,
    sequence bigint, state text, manifest_sha256 text, legal_hold boolean,
    database_system_id uuid, protected_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT r.recovery_set_id, r.operation_id, r.environment_id,
           r.sequence, r.state, r.manifest_sha256, r.legal_hold,
           r.database_system_id, r.protected_at
      FROM platform.platformops_recovery_sets r
     WHERE r.environment_id = p_environment_id AND r.state <> 'deleted'
     ORDER BY r.sequence DESC, r.recovery_set_id DESC;
END;
$function$;

CREATE FUNCTION platform.platformops_request_restore_v2(
    p_operation_id uuid, p_recovery_set_id uuid, p_environment_id uuid,
    p_target_disposition text, p_plan_sha256 text, p_plan_bytes bytea,
    p_safety_recovery_set_id uuid, p_safety_recovery_set_sha256 text,
    p_waiver_reason text, p_probe_sha256 text,
    p_destructive_confirmed boolean, p_created_at timestamptz,
    p_tenant_id uuid, p_actor_user_id uuid, p_session_id uuid
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM platform.tenancy_tenant_bindings b
         WHERE b.environment_id = p_environment_id
           AND b.tenant_id = p_tenant_id
           AND platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id)
    ) THEN
        RAISE EXCEPTION 'current owner is required' USING ERRCODE = '42501';
    END IF;
    PERFORM platform.platformops_request_restore_v1(
        p_operation_id, p_recovery_set_id, p_environment_id,
        p_target_disposition, p_plan_sha256, p_plan_bytes,
        p_safety_recovery_set_id, p_safety_recovery_set_sha256,
        p_waiver_reason, p_probe_sha256, p_destructive_confirmed, p_created_at
    );
    UPDATE platform.platformops_restore_operations r
       SET requested_by_actor_id = p_actor_user_id,
           requested_by_session_id = p_session_id
     WHERE r.operation_id = p_operation_id
       AND (r.requested_by_actor_id IS NULL
            OR (r.requested_by_actor_id = p_actor_user_id
                AND r.requested_by_session_id = p_session_id));
    IF NOT FOUND THEN
        RAISE EXCEPTION 'restore requester conflicts' USING ERRCODE = '23505';
    END IF;
END;
$function$;

CREATE FUNCTION platform.platformops_claim_recovery_manifest_v1(
    p_recovery_set_id uuid, p_tenant_id uuid, p_actor_user_id uuid,
    p_session_id uuid, p_claimed_at timestamptz
)
RETURNS TABLE(artifact_sha256 text, artifact_bytes bytea)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION 'current owner is required' USING ERRCODE = '42501';
    END IF;
    INSERT INTO platform.platformops_recovery_artifact_claims(
        artifact_kind, artifact_id, tenant_id, actor_user_id, session_id,
        artifact_sha256, claimed_at
    )
    SELECT 'recovery_manifest', r.recovery_set_id, p_tenant_id,
           p_actor_user_id, p_session_id, r.manifest_sha256, p_claimed_at
      FROM platform.platformops_recovery_sets r
      JOIN platform.tenancy_tenant_bindings b
        ON b.environment_id = r.environment_id AND b.tenant_id = p_tenant_id
     WHERE r.recovery_set_id = p_recovery_set_id
       AND r.state = 'protected' AND r.manifest_bytes IS NOT NULL
    ON CONFLICT DO NOTHING;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    RETURN QUERY
    SELECT r.manifest_sha256, r.manifest_bytes
      FROM platform.platformops_recovery_sets r
     WHERE r.recovery_set_id = p_recovery_set_id;
END;
$function$;

CREATE FUNCTION platform.platformops_claim_restore_plan_v1(
    p_operation_id uuid, p_tenant_id uuid, p_actor_user_id uuid,
    p_session_id uuid, p_claimed_at timestamptz
)
RETURNS TABLE(artifact_sha256 text, artifact_bytes bytea)
LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION 'current owner is required' USING ERRCODE = '42501';
    END IF;
    INSERT INTO platform.platformops_recovery_artifact_claims(
        artifact_kind, artifact_id, tenant_id, actor_user_id, session_id,
        artifact_sha256, claimed_at
    )
    SELECT 'restore_plan', r.operation_id, p_tenant_id,
           p_actor_user_id, p_session_id, r.plan_sha256, p_claimed_at
      FROM platform.platformops_restore_operations r
      JOIN platform.tenancy_tenant_bindings b
        ON b.environment_id = r.environment_id AND b.tenant_id = p_tenant_id
     WHERE r.operation_id = p_operation_id AND r.state = 'plan_ready'
       AND r.requested_by_actor_id = p_actor_user_id
       AND r.requested_by_session_id = p_session_id
       AND r.created_at + interval '15 minutes' > p_claimed_at
    ON CONFLICT DO NOTHING;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    RETURN QUERY
    SELECT r.plan_sha256, r.plan_bytes
      FROM platform.platformops_restore_operations r
     WHERE r.operation_id = p_operation_id;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platformops_list_recovery_sets_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_request_restore_v2(
    uuid,uuid,uuid,text,text,bytea,uuid,text,text,text,boolean,timestamptz,
    uuid,uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_claim_recovery_manifest_v1(
    uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_claim_restore_plan_v1(
    uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_list_recovery_sets_v1(uuid)
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_request_restore_v2(
    uuid,uuid,uuid,text,text,bytea,uuid,text,text,text,boolean,timestamptz,
    uuid,uuid,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_recovery_manifest_v1(
    uuid,uuid,uuid,uuid,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_restore_plan_v1(
    uuid,uuid,uuid,uuid,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
