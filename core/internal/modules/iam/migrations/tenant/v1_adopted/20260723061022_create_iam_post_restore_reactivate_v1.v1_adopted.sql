-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_post_restore_owner_context_v1(p_owner_user_id uuid)
RETURNS TABLE(identifier_id uuid,normalized_value text,normalized_sha256 bytea,revision bigint)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
BEGIN
RETURN QUERY
SELECT identifier.id,identifier.normalized_value,identifier.normalized_sha256,identifier.revision
FROM iam_user_identifiers AS identifier
JOIN iam_users AS subject ON subject.id=identifier.user_id
WHERE identifier.user_id=p_owner_user_id AND identifier.kind='email_v1'
  AND identifier.state='active' AND subject.status='active';
END;
$function$;

CREATE FUNCTION iam_post_restore_reactivate_owner_v1(
    p_operation_id uuid,p_environment_id uuid,p_restore_operation_id uuid,p_tenant_id uuid,
    p_owner_user_id uuid,p_anchor_id uuid,p_assertion_sha256 bytea,
    p_password_credential_sha256 bytea,p_receipt_sha256 bytea,p_sign_count bigint,
    p_salt bytea,p_tag bytea,p_memory_kib integer,p_iterations smallint,p_parallelism smallint,
    p_identifier_id uuid,p_identifier_revision bigint,p_directory_key_id text,
    p_directory_hmac bytea,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE v_existing iam_post_restore_owner_receipts%ROWTYPE;
v_identifier iam_user_identifiers%ROWTYPE;
v_current_version bigint;
BEGIN
    IF p_operation_id IS NULL OR p_environment_id IS NULL OR p_restore_operation_id IS NULL OR
       p_tenant_id IS NULL OR p_owner_user_id IS NULL OR p_anchor_id IS NULL OR
       pg_catalog.octet_length(p_assertion_sha256)<>32 OR
       pg_catalog.octet_length(p_password_credential_sha256)<>32 OR
       pg_catalog.octet_length(p_receipt_sha256)<>32 OR p_sign_count<0 OR
       pg_catalog.octet_length(p_salt)<>16 OR pg_catalog.octet_length(p_tag)<>32 OR
       p_memory_kib<>65536 OR p_iterations<>3 OR p_parallelism<>4 OR
       p_identifier_id IS NULL OR p_identifier_revision<1 OR
       p_directory_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_directory_hmac)<>32 THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='post-restore owner result is invalid';
    END IF;
    SELECT * INTO v_existing FROM iam_post_restore_owner_receipts
     WHERE restore_operation_id=p_restore_operation_id;
    IF FOUND THEN
        IF v_existing.operation_id<>p_operation_id OR v_existing.owner_user_id<>p_owner_user_id OR
           v_existing.anchor_id<>p_anchor_id OR v_existing.assertion_sha256<>p_assertion_sha256 OR
           v_existing.password_credential_sha256<>p_password_credential_sha256 OR
           v_existing.receipt_sha256<>p_receipt_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',MESSAGE='post-restore owner replay conflicts';
        END IF;
        RETURN false;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM iam_restored_identity_receipts
        WHERE restore_operation_id=p_restore_operation_id AND environment_id=p_environment_id
          AND owner_user_id=p_owner_user_id AND anchor_id=p_anchor_id) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='restored identity receipt is unavailable';
    END IF;
    SELECT * INTO v_identifier FROM iam_user_identifiers
     WHERE id=p_identifier_id AND user_id=p_owner_user_id AND kind='email_v1'
       AND state='active' AND revision=p_identifier_revision FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='restored owner identifier is unavailable';
    END IF;
    SELECT credential_version INTO v_current_version FROM iam_password_credentials
     WHERE user_id=p_owner_user_id FOR UPDATE;
    IF FOUND THEN
        UPDATE iam_password_credentials SET credential_version=v_current_version+1,state='active',
            algorithm='argon2id_v1',salt=p_salt,tag=p_tag,memory_kib=p_memory_kib,
            iterations=p_iterations,parallelism=p_parallelism,legacy_bcrypt_hash=NULL,
            upgrade_required=false,revision=revision+1,updated_at=p_occurred_at,
            last_used_at=NULL,invalidated_at=NULL WHERE user_id=p_owner_user_id;
    ELSE
        INSERT INTO iam_password_credentials(user_id,credential_version,state,algorithm,salt,tag,
            memory_kib,iterations,parallelism,legacy_bcrypt_hash,upgrade_required,revision,
            created_at,updated_at,last_used_at,invalidated_at)
        VALUES(p_owner_user_id,1,'active','argon2id_v1',p_salt,p_tag,p_memory_kib,p_iterations,
            p_parallelism,NULL,false,1,p_occurred_at,p_occurred_at,NULL,NULL);
    END IF;
    UPDATE iam_users SET token_version=token_version+1,authorization_version=authorization_version+1,
        revision=revision+1,updated_at=p_occurred_at WHERE id=p_owner_user_id AND status='active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='restored owner is unavailable';
    END IF;
    IF NOT platform.iam_post_restore_reactivation_apply_v1(
        p_operation_id,p_environment_id,p_restore_operation_id,p_tenant_id,
        p_owner_user_id,p_anchor_id,p_assertion_sha256,
        p_password_credential_sha256,p_receipt_sha256,p_sign_count,
        p_identifier_id,p_identifier_revision,p_directory_key_id,
        p_directory_hmac,v_identifier.normalized_sha256,p_occurred_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='post-restore owner completion replayed';
    END IF;
    INSERT INTO iam_post_restore_owner_receipts VALUES(p_restore_operation_id,p_operation_id,
        p_owner_user_id,p_anchor_id,p_assertion_sha256,p_password_credential_sha256,
        p_receipt_sha256,p_occurred_at);
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_post_restore_owner_context_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_post_restore_reactivate_owner_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,bytea,bytea,
    integer,smallint,smallint,uuid,bigint,text,bytea,timestamptz
) FROM PUBLIC;
