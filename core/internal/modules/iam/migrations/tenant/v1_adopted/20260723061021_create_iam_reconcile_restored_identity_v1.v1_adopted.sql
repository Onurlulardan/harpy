-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reconcile_restored_identity_v1(
    p_restore_operation_id uuid,p_environment_id uuid,p_recovery_set_id uuid,
    p_owner_user_id uuid,p_anchor_id uuid,p_security_epoch bytea,
    p_restore_receipt_sha256 bytea,p_protected_head_sha256 bytea,
    p_custody_generation bigint,p_custody_result_sha256 bytea,p_occurred_at timestamptz
)
RETURNS TABLE(invalidation_counts jsonb,invalidation_sha256 bytea,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE v_existing iam_restored_identity_receipts%ROWTYPE;
v_sessions bigint; v_families bigint; v_tokens bigint;
v_passwords bigint; v_authenticators bigint; v_actions bigint;
v_grants bigint; v_counts jsonb; v_digest bytea;
BEGIN
    v_sessions:=0; v_families:=0; v_tokens:=0; v_passwords:=0;
    v_authenticators:=0; v_actions:=0; v_grants:=0;
    IF p_restore_operation_id IS NULL OR p_environment_id IS NULL OR p_recovery_set_id IS NULL OR
       p_owner_user_id IS NULL OR p_anchor_id IS NULL OR pg_catalog.octet_length(p_security_epoch)<>32 OR
       pg_catalog.octet_length(p_restore_receipt_sha256)<>32 OR
       pg_catalog.octet_length(p_protected_head_sha256)<>32 OR p_custody_generation<1 OR
       pg_catalog.octet_length(p_custody_result_sha256)<>32 OR NOT EXISTS (
           SELECT 1 FROM iam_users WHERE id=p_owner_user_id AND status='active'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='restored IAM identity input is invalid';
    END IF;
    SELECT * INTO v_existing FROM iam_restored_identity_receipts
     WHERE restore_operation_id=p_restore_operation_id;
    IF FOUND THEN
        IF v_existing.environment_id<>p_environment_id OR
           v_existing.recovery_set_id<>p_recovery_set_id OR
           v_existing.owner_user_id<>p_owner_user_id OR v_existing.anchor_id<>p_anchor_id OR
           v_existing.security_epoch<>p_security_epoch OR
           v_existing.restore_receipt_sha256<>p_restore_receipt_sha256 OR
           v_existing.protected_head_sha256<>p_protected_head_sha256 OR
           v_existing.custody_generation<>p_custody_generation OR
           v_existing.custody_result_sha256<>p_custody_result_sha256 THEN
		    RETURN QUERY SELECT v_existing.invalidation_counts,
		        v_existing.invalidation_sha256,false;
		    RETURN;
        END IF;
        RETURN QUERY SELECT v_existing.invalidation_counts,v_existing.invalidation_sha256,true;
        RETURN;
    END IF;
    UPDATE iam_sessions SET state='revoked',revision=revision+1,updated_at=p_occurred_at,
        revoked_at=p_occurred_at,retain_until=GREATEST(retain_until,p_occurred_at+INTERVAL '30 days')
     WHERE state='active'; GET DIAGNOSTICS v_sessions=ROW_COUNT;
    UPDATE iam_refresh_families SET state='revoked',revoke_reason='restore_recovery',
        revision=revision+1,updated_at=p_occurred_at,revoked_at=p_occurred_at
     WHERE state='active'; GET DIAGNOSTICS v_families=ROW_COUNT;
    UPDATE iam_refresh_tokens SET state='revoked',revision=revision+1,
        updated_at=p_occurred_at,revoked_at=p_occurred_at
     WHERE state='active'; GET DIAGNOSTICS v_tokens=ROW_COUNT;
    UPDATE iam_password_credentials SET state='disabled',credential_version=credential_version+1,
        revision=revision+1,updated_at=p_occurred_at,invalidated_at=p_occurred_at
     WHERE state IN ('active','reset_required'); GET DIAGNOSTICS v_passwords=ROW_COUNT;
    UPDATE iam_authenticators SET state='renewal_required',revision=revision+1,
        updated_at=p_occurred_at,renewal_required_at=p_occurred_at
     WHERE state='active'; GET DIAGNOSTICS v_authenticators=ROW_COUNT;
    UPDATE iam_recovery_codes SET state='revoked',revision=revision+1,updated_at=p_occurred_at,
        revoked_at=p_occurred_at,purge_after=p_occurred_at+INTERVAL '30 days' WHERE state='active';
    UPDATE iam_auth_challenges SET state='cancelled',revision=revision+1,updated_at=p_occurred_at,
        terminal_at=p_occurred_at,purge_after=GREATEST(purge_after,p_occurred_at+INTERVAL '24 hours')
     WHERE state='pending';
    UPDATE iam_one_time_tokens SET state='revoked',revision=revision+1,updated_at=p_occurred_at,
        revoked_at=p_occurred_at,purge_after=GREATEST(purge_after,p_occurred_at+INTERVAL '24 hours')
     WHERE state='pending'; GET DIAGNOSTICS v_actions=ROW_COUNT;
    UPDATE iam_invitations SET state='revoked',revision=revision+1,updated_at=p_occurred_at,
        revoked_at=p_occurred_at WHERE state='pending';
    UPDATE iam_initial_admin_setups SET state='revoked',revision=revision+1,revoked_at=p_occurred_at
     WHERE state='pending';
    UPDATE iam_user_role_assignments SET state='suspended',revision=revision+1,updated_at=p_occurred_at
     WHERE state='active' AND user_id<>p_owner_user_id; GET DIAGNOSTICS v_grants=ROW_COUNT;
    UPDATE iam_user_permission_bindings SET state='revoked',revision=revision+1,
        updated_at=p_occurred_at,revoked_at=p_occurred_at
     WHERE state='active' AND user_id<>p_owner_user_id;
    UPDATE iam_role_permission_bindings AS binding SET state='revoked',revision=binding.revision+1,
        updated_at=p_occurred_at,revoked_at=p_occurred_at FROM iam_roles AS role
     WHERE binding.role_id=role.id AND binding.state='active' AND NOT role.is_protected;
    UPDATE iam_organization_permission_bindings SET state='revoked',revision=revision+1,
        updated_at=p_occurred_at,revoked_at=p_occurred_at WHERE state='active';
    UPDATE iam_capability_grants SET state='revoked',revision=revision+1,
        updated_at=p_occurred_at,revoked_at=p_occurred_at WHERE state='active';
    UPDATE iam_users SET token_version=token_version+1,authorization_version=authorization_version+1,
        revision=revision+1,updated_at=p_occurred_at;
    UPDATE iam_authorization_clock SET clock_value=clock_value+1,revision=revision+1,
        updated_at=p_occurred_at WHERE singleton;
    v_counts:=pg_catalog.jsonb_build_object('sessions',v_sessions,'refresh_families',v_families,
        'refresh_tokens',v_tokens,'passwords',v_passwords,'authenticators',v_authenticators,
        'action_tokens',v_actions,'delegable_grants',v_grants);
    v_digest:=pg_catalog.sha256(pg_catalog.convert_to(v_counts::text,'UTF8'));
    INSERT INTO iam_restored_identity_receipts VALUES (p_restore_operation_id,p_environment_id,
        p_recovery_set_id,p_owner_user_id,p_anchor_id,p_security_epoch,p_restore_receipt_sha256,
        p_protected_head_sha256,p_custody_generation,p_custody_result_sha256,v_counts,v_digest,p_occurred_at);
    RETURN QUERY SELECT v_counts,v_digest,false;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_reconcile_restored_identity_v1(
    uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,bytea,timestamptz
) FROM PUBLIC;
