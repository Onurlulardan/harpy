-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_purge_data_v2(
    p_operation_id uuid,p_component_id uuid,p_generation bigint,
    p_content_sha256 text,p_reference_epoch bigint,
    p_completed_at timestamptz
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component iam_catalog_components%ROWTYPE;
    v_head iam_catalog_component_heads%ROWTYPE;
    v_blockers jsonb;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_generation<1 OR p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_reference_epoch<1 OR p_completed_at IS NULL OR
       NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM Catalog purge data input is invalid';
    END IF;
    SELECT * INTO v_component
    FROM iam_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    SELECT * INTO v_head
    FROM iam_catalog_component_heads AS head
    WHERE head.component_id=p_component_id FOR UPDATE;
    IF v_component.component_id IS NULL OR v_head.component_id IS NULL OR
       v_component.lifecycle_state<>'RETIRED' OR
       v_component.owner_generation<>p_generation OR
       v_component.content_sha256<>p_content_sha256 OR
       v_head.owner_generation<>p_generation OR
       v_head.content_sha256<>p_content_sha256 OR
       p_completed_at<v_component.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog purge data authority differs';
    END IF;
    PERFORM permission.id
    FROM iam_permissions AS permission
    WHERE permission.id IN (
            SELECT mapped.permission_id
            FROM iam_catalog_version_operational_grants AS mapped
            WHERE mapped.component_id=p_component_id
        ) OR permission.resource_id IN (
            SELECT mapped.resource_id
            FROM iam_catalog_version_operational_resources AS mapped
            WHERE mapped.component_id=p_component_id
        )
    ORDER BY permission.id FOR UPDATE OF permission;
    PERFORM resource.id
    FROM iam_resources AS resource
    WHERE resource.id IN (
        SELECT mapped.resource_id
        FROM iam_catalog_version_operational_resources AS mapped
        WHERE mapped.component_id=p_component_id
    ) ORDER BY resource.id FOR UPDATE OF resource;
    PERFORM binding.id
    FROM iam_role_permission_bindings AS binding
    WHERE binding.permission_id IN (
        SELECT mapped.permission_id
        FROM iam_catalog_version_operational_grants AS mapped
        WHERE mapped.component_id=p_component_id
    ) ORDER BY binding.id FOR UPDATE OF binding;
    PERFORM binding.id
    FROM iam_user_permission_bindings AS binding
    WHERE binding.permission_id IN (
        SELECT mapped.permission_id
        FROM iam_catalog_version_operational_grants AS mapped
        WHERE mapped.component_id=p_component_id
    ) ORDER BY binding.id FOR UPDATE OF binding;
    PERFORM binding.id
    FROM iam_organization_permission_bindings AS binding
    WHERE binding.permission_id IN (
        SELECT mapped.permission_id
        FROM iam_catalog_version_operational_grants AS mapped
        WHERE mapped.component_id=p_component_id
    ) ORDER BY binding.id FOR UPDATE OF binding;
    SELECT snapshot.blockers INTO STRICT v_blockers
    FROM iam_catalog_blocker_snapshot_v2(
        p_component_id,p_reference_epoch,p_generation
    ) AS snapshot;
    IF v_blockers<>'[]'::jsonb OR EXISTS (
        SELECT 1 FROM iam_permissions AS permission
        WHERE permission.id IN (
            SELECT mapped.permission_id
            FROM iam_catalog_version_operational_grants AS mapped
            WHERE mapped.component_id=p_component_id
        ) AND permission.updated_at>p_completed_at
    ) OR EXISTS (
        SELECT 1 FROM iam_resources AS resource
        WHERE resource.id IN (
            SELECT mapped.resource_id
            FROM iam_catalog_version_operational_resources AS mapped
            WHERE mapped.component_id=p_component_id
        ) AND resource.updated_at>p_completed_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog purge data is blocked or stale';
    END IF;
    UPDATE iam_permissions AS permission
    SET state='retired',revision=permission.revision+1,
        updated_at=p_completed_at,retired_at=p_completed_at
    WHERE permission.state='active' AND permission.id IN (
        SELECT mapped.permission_id
        FROM iam_catalog_version_operational_grants AS mapped
        WHERE mapped.component_id=p_component_id
    );
    UPDATE iam_resources AS resource
    SET state='retired',revision=resource.revision+1,
        updated_at=p_completed_at,retired_at=p_completed_at
    WHERE resource.state='active' AND resource.id IN (
        SELECT mapped.resource_id
        FROM iam_catalog_version_operational_resources AS mapped
        WHERE mapped.component_id=p_component_id
    );
    INSERT INTO iam_catalog_purge_contexts(
        backend_pid,transaction_id,operation_id,component_id,opened_at
    ) VALUES (
        pg_catalog.pg_backend_pid(),pg_catalog.pg_current_xact_id(),
        p_operation_id,p_component_id,p_completed_at
    );
    DELETE FROM iam_catalog_component_candidates
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_component_heads
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_version_operational_grants
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_version_operational_resources
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_version_provisions
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_version_grants
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_version_resources
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_version_evidence
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_operational_grants
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_operational_resources
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_provisions
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_grant_projections
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_resource_projections
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_materializations
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_component_blockers
    WHERE component_id=p_component_id;
    DELETE FROM iam_catalog_purge_contexts AS context
    WHERE context.backend_pid=pg_catalog.pg_backend_pid()
      AND context.transaction_id=pg_catalog.pg_current_xact_id()
      AND context.component_id=p_component_id;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_purge_data_v2(
    uuid,uuid,bigint,text,bigint,timestamptz
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_catalog_purge_data_v2(
    uuid,uuid,bigint,text,bigint,timestamptz
) FROM harpy_runtime;
