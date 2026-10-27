-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_post_restore_challenge_v1(
    p_operation_id uuid,p_environment_id uuid,p_restore_operation_id uuid,
    p_challenge text,p_context_sha256 bytea,p_rp_id text,p_origin text,
    p_expires_at timestamptz,p_occurred_at timestamptz
)
RETURNS TABLE(operation_id uuid,tenant_id uuid,owner_user_id uuid,anchor_id uuid,
    challenge text,credential_id bytea,public_key_cose bytea,rp_id text,
    sign_count bigint,expires_at timestamptz,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
DECLARE v_reconcile platform.iam_post_restore_reconciliations%ROWTYPE;
v_existing platform.iam_post_restore_challenges%ROWTYPE;
v_replayed boolean := false;
BEGIN
    IF p_operation_id IS NULL OR p_environment_id IS NULL OR p_restore_operation_id IS NULL OR
       p_challenge !~ '^[A-Za-z0-9_-]{43}$' OR pg_catalog.octet_length(p_context_sha256)<>32 OR
       pg_catalog.octet_length(p_rp_id) NOT BETWEEN 1 AND 253 OR
       pg_catalog.octet_length(p_origin) NOT BETWEEN 1 AND 512 OR
       p_expires_at<=p_occurred_at OR p_expires_at>p_occurred_at+INTERVAL '5 minutes' THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='post-restore challenge input is invalid';
    END IF;
    SELECT * INTO v_reconcile FROM platform.iam_post_restore_reconciliations
     WHERE restore_operation_id=p_restore_operation_id FOR UPDATE;
    IF NOT FOUND OR v_reconcile.environment_id<>p_environment_id OR
       v_reconcile.state<>'identity_reactivation_required' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='post-restore reactivation is unavailable';
    END IF;
    UPDATE platform.iam_post_restore_challenges SET state='expired',terminal_at=p_occurred_at
     WHERE restore_operation_id=p_restore_operation_id AND state='pending'
       AND expires_at<=p_occurred_at;
	UPDATE platform.iam_post_restore_challenges SET state='failed',terminal_at=p_occurred_at
	 WHERE restore_operation_id=p_restore_operation_id AND state='pending'
	   AND operation_id<>p_operation_id;
    SELECT * INTO v_existing FROM platform.iam_post_restore_challenges
     WHERE operation_id=p_operation_id FOR SHARE;
    IF FOUND THEN
        v_replayed := true;
        IF v_existing.restore_operation_id<>p_restore_operation_id OR
           v_existing.environment_id<>p_environment_id OR v_existing.challenge<>p_challenge OR
           v_existing.context_sha256<>p_context_sha256 OR v_existing.rp_id<>p_rp_id OR
           v_existing.origin<>p_origin OR v_existing.expires_at<>p_expires_at THEN
            RAISE EXCEPTION USING ERRCODE='23505',MESSAGE='post-restore challenge replay conflicts';
        END IF;
    ELSE
        INSERT INTO platform.iam_post_restore_challenges(
            operation_id,restore_operation_id,environment_id,tenant_id,owner_user_id,
            anchor_id,challenge,context_sha256,rp_id,origin,state,created_at,expires_at
        ) VALUES (p_operation_id,p_restore_operation_id,p_environment_id,
            v_reconcile.tenant_id,v_reconcile.owner_user_id,v_reconcile.anchor_id,
            p_challenge,p_context_sha256,p_rp_id,p_origin,'pending',p_occurred_at,p_expires_at);
        SELECT * INTO v_existing FROM platform.iam_post_restore_challenges
         WHERE operation_id=p_operation_id;
    END IF;
    RETURN QUERY SELECT v_existing.operation_id,v_existing.tenant_id,v_existing.owner_user_id,
        v_existing.anchor_id,v_existing.challenge,anchor.credential_id,anchor.public_key_cose,
        v_existing.rp_id,anchor.sign_count,v_existing.expires_at,v_replayed
      FROM platform.iam_recovery_anchors AS anchor
     WHERE anchor.anchor_id=v_existing.anchor_id AND anchor.state='active'
       AND anchor.user_verification='required' AND NOT anchor.backup_eligible
       AND NOT anchor.backup_state;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='post-restore recovery anchor is unavailable';
    END IF;
END;
$function$;

CREATE FUNCTION platform.iam_post_restore_challenge_context_v1(
    p_operation_id uuid,p_environment_id uuid,p_restore_operation_id uuid,p_origin text,p_occurred_at timestamptz
)
RETURNS TABLE(tenant_id uuid,owner_user_id uuid,anchor_id uuid,challenge text,
    credential_id bytea,public_key_cose bytea,rp_id text,sign_count bigint,
	context_sha256 bytea)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
BEGIN
    RETURN QUERY SELECT challenge.tenant_id,challenge.owner_user_id,challenge.anchor_id,
        challenge.challenge,anchor.credential_id,anchor.public_key_cose,
		challenge.rp_id,anchor.sign_count,challenge.context_sha256
      FROM platform.iam_post_restore_challenges AS challenge
      JOIN platform.iam_post_restore_reconciliations AS reconciliation
        ON reconciliation.restore_operation_id=challenge.restore_operation_id
      JOIN platform.iam_recovery_anchors AS anchor ON anchor.anchor_id=challenge.anchor_id
     WHERE challenge.operation_id=p_operation_id
       AND challenge.environment_id=p_environment_id
       AND challenge.restore_operation_id=p_restore_operation_id
       AND challenge.origin=p_origin AND challenge.state='pending'
       AND challenge.expires_at>p_occurred_at
       AND reconciliation.state='identity_reactivation_required'
       AND anchor.state='active' AND anchor.user_verification='required'
       AND NOT anchor.backup_eligible AND NOT anchor.backup_state;
END;
$function$;

CREATE FUNCTION platform.iam_post_restore_binding_v1(
    p_environment_id uuid,p_restore_operation_id uuid
)
RETURNS TABLE(tenant_id uuid,owner_user_id uuid,anchor_id uuid,
    new_database_system_id uuid,security_epoch bytea,tenant_slug text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
BEGIN
RETURN QUERY
SELECT reconciliation.tenant_id,reconciliation.owner_user_id,reconciliation.anchor_id,
       reconciliation.new_database_system_id,reconciliation.security_epoch,tenant.slug::text
FROM platform.iam_post_restore_reconciliations AS reconciliation
	JOIN platform.tenants AS tenant ON tenant.id=reconciliation.tenant_id
WHERE reconciliation.environment_id=p_environment_id
  AND reconciliation.restore_operation_id=p_restore_operation_id
  AND reconciliation.state='identity_reactivation_required';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_post_restore_challenge_v1(
    uuid,uuid,uuid,text,bytea,text,text,timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_post_restore_challenge_context_v1(
    uuid,uuid,uuid,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_post_restore_binding_v1(uuid,uuid) FROM PUBLIC;
