-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_capability_grant_write_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_grant_id uuid,
    p_principal_id uuid,
    p_expected_revision bigint,
    p_capability_key text,
    p_scope_kind text,
    p_owner_domain text,
    p_resource_kind text,
    p_resource_id uuid,
    p_valid_from timestamptz,
    p_valid_until timestamptz,
    p_manifest_sha256 bytea,
    p_actor_id uuid,
    p_manage_permission_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_purpose_sha256 bytea,
    p_auth_time timestamptz,
    p_amr_password boolean,
    p_amr_webauthn boolean,
    p_amr_totp boolean
)
RETURNS TABLE (
    result_id uuid,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz;
    v_new_revision bigint;
    v_head_principal uuid;
    v_head_capability text;
    v_head_revision bigint;
    v_head_state text;
    v_receipt_command text;
    v_receipt_request bytea;
    v_receipt_id uuid;
    v_receipt_revision bigint;
    v_receipt_clock bigint;
    v_clock bigint;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    v_new_revision := p_expected_revision + 1;
    IF p_operation_id IS NULL OR p_grant_id IS NULL OR p_principal_id IS NULL OR
       p_expected_revision < 0 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_manifest_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM capability grant input is invalid';
    END IF;
    IF NOT iam_service_principal_actor_assert_v1(
        p_scope_schema, p_actor_id, p_manage_permission_id,
        p_expected_token_version, p_expected_authorization_version,
        p_purpose_sha256, p_auth_time,
        p_amr_password, p_amr_webauthn, p_amr_totp
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM service principal command authority is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
           receipt.result_revision, receipt.clock_value
    INTO v_receipt_command, v_receipt_request, v_receipt_id,
         v_receipt_revision, v_receipt_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id OPERATOR(pg_catalog.=) p_operation_id;
    IF FOUND THEN
        IF v_receipt_command IS DISTINCT FROM 'iam.capability_grant.write' OR
           v_receipt_request IS DISTINCT FROM p_request_sha256 OR
           v_receipt_id IS DISTINCT FROM p_grant_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM capability grant operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_receipt_id, v_receipt_revision,
            v_receipt_clock, true;
        RETURN;
    END IF;
    PERFORM principal.id
    FROM iam_service_principals AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_principal_id
      AND principal.state OPERATOR(pg_catalog.=) 'active'
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM capability grant principal is inactive';
    END IF;
    SELECT grant_head.principal_id, grant_head.capability_key,
           grant_head.current_revision, grant_head.state
    INTO v_head_principal, v_head_capability, v_head_revision, v_head_state
    FROM iam_capability_grants AS grant_head
    WHERE grant_head.id OPERATOR(pg_catalog.=) p_grant_id
    FOR UPDATE;
    IF p_expected_revision = 0 THEN
        IF FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM capability grant create CAS is stale';
        END IF;
        INSERT INTO iam_capability_grants (
            id, principal_id, capability_key, current_revision,
            state, revision, created_at, updated_at
        ) VALUES (
            p_grant_id, p_principal_id, p_capability_key, 1,
            'active', 1, v_now, v_now
        );
    ELSIF NOT FOUND OR v_head_principal IS DISTINCT FROM p_principal_id OR
          v_head_capability IS DISTINCT FROM p_capability_key OR
          v_head_revision IS DISTINCT FROM p_expected_revision OR
          v_head_state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM capability grant revision CAS is stale';
    END IF;
    INSERT INTO iam_capability_grant_revisions (
        grant_id, grant_revision, previous_revision, principal_id,
        capability_key, scope_kind, owner_domain, resource_kind,
        resource_id, valid_from, valid_until, disposition,
        manifest_sha256, created_by_user_id, created_at
    ) VALUES (
        p_grant_id, v_new_revision,
        CASE WHEN p_expected_revision = 0 THEN NULL ELSE p_expected_revision END,
        p_principal_id, p_capability_key, p_scope_kind, p_owner_domain,
        p_resource_kind, p_resource_id, p_valid_from, p_valid_until,
        'active', p_manifest_sha256, p_actor_id, v_now
    );
    IF p_expected_revision > 0 THEN
        UPDATE iam_capability_grants
        SET current_revision = v_new_revision, revision = revision + 1,
            updated_at = v_now
        WHERE id = p_grant_id;
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1, updated_at = v_now
    WHERE singleton
    RETURNING clock_value INTO v_clock;
    UPDATE iam_service_principals
    SET authorization_version = v_clock, revision = revision + 1, updated_at = v_now
    WHERE id = p_principal_id;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.capability_grant.write', p_request_sha256,
        v_clock, 'iam.capability_grant.written', p_grant_id,
        v_new_revision, p_request_sha256, v_now
    );
    RETURN QUERY SELECT p_grant_id, v_new_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_capability_grant_write_v1(
    name,uuid,bytea,uuid,uuid,bigint,text,text,text,text,uuid,
    timestamptz,timestamptz,bytea,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) FROM PUBLIC;
