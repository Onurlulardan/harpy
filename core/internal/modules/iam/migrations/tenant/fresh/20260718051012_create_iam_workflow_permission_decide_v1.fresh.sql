-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_workflow_permission_decide_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_request_id uuid,
    p_decision text,
    p_principal_id uuid,
    p_principal_revision bigint,
    p_principal_authorization_version bigint,
    p_grant_id uuid,
    p_grant_revision bigint,
    p_definition_id uuid,
    p_manifest_sha256 bytea,
    p_effect_key text,
    p_plan_sha256 bytea,
    p_verified_at timestamptz
)
RETURNS TABLE (receipt_document jsonb, replayed boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_group_id uuid;
    v_group_revision bigint;
    v_group_state text;
    v_group_expires_at timestamptz;
    v_request_revision bigint;
    v_request_state text;
    v_request uuid;
    v_binding uuid;
    v_bumped integer;
    v_clock bigint;
    v_result_code text;
    v_document jsonb;
BEGIN
    IF p_operation_id IS NULL OR p_request_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_decision NOT IN ('APPROVE','REJECT') OR
       pg_catalog.char_length(p_effect_key) NOT BETWEEN 1 AND 256 OR
       p_effect_key !~ '^[A-Za-z0-9][A-Za-z0-9_.:/-]*$' OR
       pg_catalog.octet_length(p_plan_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Workflow permission decision input is invalid';
    END IF;
    PERFORM iam_workflow_permission_scope_assert_v1(
        p_scope_schema, p_tenant_id
    );
    PERFORM iam_workflow_permission_authority_assert_v1(
        p_principal_id, p_principal_revision,
        p_principal_authorization_version, p_grant_id, p_grant_revision,
        p_definition_id, p_manifest_sha256, p_verified_at
    );
    SELECT receipt.receipt_document INTO v_document
    FROM iam_workflow_permission_decision_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id
      AND receipt.effect_key=p_effect_key
      AND receipt.request_sha256=p_request_sha256
      AND receipt.request_id=p_request_id AND receipt.decision=p_decision
      AND receipt.principal_id=p_principal_id
      AND receipt.principal_revision=p_principal_revision
      AND receipt.principal_authorization_version=
          p_principal_authorization_version
      AND receipt.grant_id=p_grant_id
      AND receipt.grant_revision=p_grant_revision
      AND receipt.definition_id=p_definition_id
      AND receipt.manifest_sha256=p_manifest_sha256
      AND receipt.plan_sha256=p_plan_sha256;
    IF FOUND THEN
        RETURN QUERY SELECT v_document, true;
        RETURN;
    END IF;
    IF EXISTS (
        SELECT 1 FROM iam_workflow_permission_decision_receipts AS receipt
        WHERE receipt.operation_id=p_operation_id OR receipt.effect_key=p_effect_key
    ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='Workflow permission effect identity was reused';
    END IF;
    SELECT request.group_id, request.revision, request.state,
           request_group.revision, request_group.state, request_group.expires_at
    INTO v_group_id, v_request_revision, v_request_state,
         v_group_revision, v_group_state, v_group_expires_at
    FROM iam_permission_requests AS request
    JOIN iam_permission_request_groups AS request_group
      ON request_group.id=request.group_id
    WHERE request.id=p_request_id
    FOR UPDATE OF request_group, request;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='Workflow permission request is unknown';
    END IF;
    IF v_group_state IS DISTINCT FROM 'IN_REVIEW' OR
       v_request_state IS DISTINCT FROM 'IN_REVIEW' OR
       v_now>=v_group_expires_at THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Workflow permission request decision CAS is stale';
    END IF;
    PERFORM request.id FROM iam_permission_requests AS request
    WHERE request.group_id=v_group_id ORDER BY request.id FOR UPDATE;
    IF EXISTS (
        SELECT 1 FROM iam_permission_requests AS request
        WHERE request.group_id=v_group_id AND request.state<>'IN_REVIEW'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Workflow permission request group CAS is stale';
    END IF;
    IF p_decision='APPROVE' THEN
        FOR v_request IN
            SELECT request.id FROM iam_permission_requests AS request
            WHERE request.group_id=v_group_id ORDER BY request.ordinal
        LOOP
            v_binding := iam_permission_request_grant_v1(v_request,v_now);
        END LOOP;
        v_bumped := iam_permission_request_bump_users_v1(v_group_id,v_now);
        v_result_code := 'iam.permission_request.approved';
    ELSE
        v_result_code := 'iam.permission_request.rejected';
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value=clock_value+1, revision=revision+1, updated_at=v_now
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM authorization clock is unavailable';
    END IF;
    UPDATE iam_permission_request_groups
    SET state=CASE WHEN p_decision='APPROVE' THEN 'APPROVED' ELSE 'REJECTED' END,
        terminal_at=v_now, revision=v_group_revision+1, updated_at=v_now
    WHERE id=v_group_id;
    UPDATE iam_permission_requests
    SET state=CASE WHEN p_decision='APPROVE' THEN 'APPROVED' ELSE 'REJECTED' END,
        granted_binding_id=CASE WHEN p_decision='APPROVE'
            THEN planned_binding_id ELSE NULL END,
        terminal_at=v_now, revision=revision+1, updated_at=v_now
    WHERE group_id=v_group_id;
    v_document := pg_catalog.jsonb_build_object(
        'contract_version','iam-workflow-permission-decision-receipt.v1',
        'receipt_id',p_operation_id,'effect_key',p_effect_key,
        'plan_sha256',pg_catalog.encode(p_plan_sha256,'hex'),
        'request_id',p_request_id,'request_revision_before',v_request_revision,
        'request_revision_after',v_request_revision+1,'group_id',v_group_id,
        'group_revision_before',v_group_revision,
        'group_revision_after',v_group_revision+1,'decision',p_decision,
        'service_principal_id',p_principal_id,
        'service_principal_revision',p_principal_revision,
        'principal_authorization_version',p_principal_authorization_version,
        'grant_id',p_grant_id,'grant_revision',p_grant_revision,
        'definition_id',p_definition_id,
        'capability_manifest_sha256',pg_catalog.encode(p_manifest_sha256,'hex'),
        'verified_at',p_verified_at,'decided_at',v_now,
        'authorization_clock',v_clock,'result_code',v_result_code
    );
    INSERT INTO iam_authorization_operation_receipts (
        operation_id,command_key,request_sha256,clock_value,result_code,
        result_id,result_revision,result_sha256,created_at
    ) VALUES (
        p_operation_id,'iam.permission_request.workflow_decide',p_request_sha256,
        v_clock,v_result_code,v_group_id,v_group_revision+1,p_request_sha256,v_now
    );
    INSERT INTO iam_workflow_permission_decision_receipts (
        operation_id,effect_key,request_sha256,request_id,
        request_revision_before,request_revision_after,group_id,
        group_revision_before,group_revision_after,decision,principal_id,
        principal_revision,principal_authorization_version,grant_id,
        grant_revision,definition_id,manifest_sha256,plan_sha256,verified_at,
        decided_at,authorization_clock,result_code,receipt_document,created_at
    ) VALUES (
        p_operation_id,p_effect_key,p_request_sha256,p_request_id,
        v_request_revision,v_request_revision+1,v_group_id,v_group_revision,
        v_group_revision+1,p_decision,p_principal_id,p_principal_revision,
        p_principal_authorization_version,p_grant_id,p_grant_revision,
        p_definition_id,p_manifest_sha256,p_plan_sha256,p_verified_at,v_now,
        v_clock,v_result_code,v_document,v_now
    );
    RETURN QUERY SELECT v_document, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_workflow_permission_decide_v1(
    name,uuid,uuid,bytea,uuid,text,uuid,bigint,bigint,uuid,bigint,uuid,
    bytea,text,bytea,timestamptz
) FROM PUBLIC;
