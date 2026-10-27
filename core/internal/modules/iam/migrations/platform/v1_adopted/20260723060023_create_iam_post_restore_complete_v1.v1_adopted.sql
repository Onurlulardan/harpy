-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_post_restore_reactivation_fail_v1(
    p_operation_id uuid,p_environment_id uuid,p_restore_operation_id uuid,p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
DECLARE v_attempts smallint;
BEGIN
    UPDATE platform.iam_post_restore_challenges
       SET attempt_count=attempt_count+1,
           state=CASE WHEN attempt_count+1>=5 THEN 'failed' ELSE state END,
           terminal_at=CASE WHEN attempt_count+1>=5 THEN p_occurred_at ELSE terminal_at END
     WHERE operation_id=p_operation_id AND environment_id=p_environment_id
       AND restore_operation_id=p_restore_operation_id AND state='pending'
       AND expires_at>p_occurred_at
     RETURNING attempt_count INTO v_attempts;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.iam_post_restore_reactivation_complete_v1(
    p_operation_id uuid,p_environment_id uuid,p_restore_operation_id uuid,
    p_tenant_id uuid,p_owner_user_id uuid,p_anchor_id uuid,p_assertion_sha256 bytea,
    p_password_credential_sha256 bytea,p_receipt_sha256 bytea,p_sign_count bigint,
    p_occurred_at timestamptz
)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
DECLARE v_challenge platform.iam_post_restore_challenges%ROWTYPE;
v_existing platform.iam_post_restore_reactivation_receipts%ROWTYPE;
v_old_sign_count bigint;
BEGIN
    IF pg_catalog.octet_length(p_assertion_sha256)<>32 OR
       pg_catalog.octet_length(p_password_credential_sha256)<>32 OR
       pg_catalog.octet_length(p_receipt_sha256)<>32 OR p_sign_count<0 THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='post-restore reactivation result is invalid';
    END IF;
    SELECT * INTO v_existing FROM platform.iam_post_restore_reactivation_receipts
     WHERE restore_operation_id=p_restore_operation_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.operation_id<>p_operation_id OR v_existing.environment_id<>p_environment_id OR
           v_existing.tenant_id<>p_tenant_id OR v_existing.owner_user_id<>p_owner_user_id OR
           v_existing.anchor_id<>p_anchor_id OR v_existing.assertion_sha256<>p_assertion_sha256 OR
           v_existing.password_credential_sha256<>p_password_credential_sha256 OR
           v_existing.receipt_sha256<>p_receipt_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',MESSAGE='post-restore reactivation replay conflicts';
        END IF;
        RETURN false;
    END IF;
    SELECT * INTO v_challenge FROM platform.iam_post_restore_challenges
     WHERE operation_id=p_operation_id AND environment_id=p_environment_id
       AND restore_operation_id=p_restore_operation_id FOR UPDATE;
    IF NOT FOUND OR v_challenge.state<>'pending' OR v_challenge.expires_at<=p_occurred_at OR
       v_challenge.tenant_id<>p_tenant_id OR v_challenge.owner_user_id<>p_owner_user_id OR
       v_challenge.anchor_id<>p_anchor_id THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='post-restore challenge is unavailable';
    END IF;
    SELECT sign_count INTO v_old_sign_count FROM platform.iam_recovery_anchors
     WHERE anchor_id=p_anchor_id AND state='active' FOR UPDATE;
    IF NOT FOUND OR (v_old_sign_count>0 AND p_sign_count<=v_old_sign_count) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='post-restore assertion counter is invalid';
    END IF;
    UPDATE platform.iam_recovery_anchors SET sign_count=p_sign_count,
        revision=revision+1,updated_at=p_occurred_at WHERE anchor_id=p_anchor_id;
    UPDATE platform.iam_post_restore_challenges SET state='consumed',terminal_at=p_occurred_at
     WHERE operation_id=p_operation_id;
    UPDATE platform.iam_post_restore_reconciliations
       SET state='reactivated',updated_at=p_occurred_at,reactivated_at=p_occurred_at
     WHERE restore_operation_id=p_restore_operation_id
       AND state='identity_reactivation_required';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='post-restore reconcile state is invalid';
    END IF;
    INSERT INTO platform.iam_post_restore_reactivation_receipts(
        restore_operation_id,operation_id,environment_id,tenant_id,owner_user_id,
        anchor_id,assertion_sha256,password_credential_sha256,receipt_sha256,created_at
    ) VALUES (p_restore_operation_id,p_operation_id,p_environment_id,p_tenant_id,
        p_owner_user_id,p_anchor_id,p_assertion_sha256,p_password_credential_sha256,
        p_receipt_sha256,p_occurred_at);
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_post_restore_reactivation_fail_v1(
    uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_post_restore_reactivation_complete_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,timestamptz
) FROM PUBLIC;
