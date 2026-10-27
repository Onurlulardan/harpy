-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_post_restore_reactivation_apply_v1(
    p_operation_id uuid,p_environment_id uuid,p_restore_operation_id uuid,
    p_tenant_id uuid,p_owner_user_id uuid,p_anchor_id uuid,
    p_assertion_sha256 bytea,p_password_credential_sha256 bytea,
    p_receipt_sha256 bytea,p_sign_count bigint,p_identifier_id uuid,
    p_identifier_revision bigint,p_directory_key_id text,
    p_directory_hmac bytea,p_source_sha256 bytea,p_occurred_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL OR p_owner_user_id IS NULL OR
       p_identifier_id IS NULL OR p_identifier_revision<1 OR
       p_directory_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_directory_hmac)<>32 OR
       pg_catalog.octet_length(p_source_sha256)<>32 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='post-restore directory projection is denied';
    END IF;
    DELETE FROM platform.iam_login_directory
    WHERE tenant_id=p_tenant_id;
    INSERT INTO platform.iam_login_directory(
        tenant_id,identifier_kind,lookup_key_id,identifier_hmac,user_id,
        source_identifier_id,source_revision,state,source_sha256,projected_at
    ) VALUES (
        p_tenant_id,'email_v1',p_directory_key_id,p_directory_hmac,
        p_owner_user_id,p_identifier_id,p_identifier_revision,'active',
        p_source_sha256,p_occurred_at
    );
    RETURN platform.iam_post_restore_reactivation_complete_v1(
        p_operation_id,p_environment_id,p_restore_operation_id,p_tenant_id,
        p_owner_user_id,p_anchor_id,p_assertion_sha256,
        p_password_credential_sha256,p_receipt_sha256,p_sign_count,p_occurred_at
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_post_restore_reactivation_apply_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,uuid,bigint,
    text,bytea,bytea,timestamptz
) FROM PUBLIC;
