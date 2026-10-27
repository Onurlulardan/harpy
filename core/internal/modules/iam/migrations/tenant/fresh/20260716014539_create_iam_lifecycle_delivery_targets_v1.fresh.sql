-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_email_verification_delivery_target_v1(
    p_scope_schema name, p_identifier_id uuid, p_actor_id uuid,
    p_permission_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    subject_user_id uuid, resource_id uuid,
    recipient_identifier_id uuid, recipient_address text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_identity_actor_assert_v1(
           p_scope_schema, p_actor_id, p_permission_id,
           p_expected_token_version, p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM email verification delivery target is denied';
    END IF;
    RETURN QUERY
    SELECT identifier.user_id, identifier.id, identifier.id,
           identifier.normalized_value
    FROM iam_user_identifiers AS identifier
    JOIN iam_users AS subject ON subject.id = identifier.user_id
    WHERE identifier.id = p_identifier_id
      AND identifier.kind = 'email_v1'
      AND identifier.state IN ('pending', 'verified')
      AND subject.status = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM email verification target is unavailable';
    END IF;
END;
$function$;

CREATE FUNCTION iam_password_reset_delivery_target_v1(
    p_scope_schema name, p_identifier_kind text,
    p_identifier_sha256 bytea
)
RETURNS TABLE (
    subject_user_id uuid, resource_id uuid,
    recipient_identifier_id uuid, recipient_address text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_identifier_kind NOT IN ('email_v1', 'portal_v1') OR
       pg_catalog.octet_length(p_identifier_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM password reset delivery target is denied';
    END IF;
    RETURN QUERY
    SELECT target.user_id, target.id, recipient.id,
           recipient.normalized_value
    FROM iam_user_identifiers AS target
    JOIN iam_users AS subject
      ON subject.id = target.user_id AND subject.status = 'active'
    JOIN iam_user_identifiers AS recipient
      ON recipient.user_id = target.user_id
     AND recipient.kind = 'email_v1' AND recipient.state = 'active'
    WHERE target.kind = p_identifier_kind
      AND target.normalized_sha256 = p_identifier_sha256
      AND target.state = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM password reset target is unavailable';
    END IF;
END;
$function$;

CREATE FUNCTION iam_portal_setup_delivery_target_v1(
    p_scope_schema name, p_binding_id uuid, p_actor_id uuid,
    p_permission_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    subject_user_id uuid, resource_id uuid,
    recipient_identifier_id uuid, recipient_address text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT iam_identity_actor_assert_v1(
           p_scope_schema, p_actor_id, p_permission_id,
           p_expected_token_version, p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM portal setup delivery target is denied';
    END IF;
    RETURN QUERY
    SELECT setup.user_id, setup.binding_id, recipient.id,
           recipient.normalized_value
    FROM iam_portal_access_setups AS setup
    JOIN iam_portal_subject_bindings AS binding
      ON binding.id = setup.binding_id AND binding.user_id = setup.user_id
    JOIN iam_users AS subject
      ON subject.id = setup.user_id AND subject.status = 'active'
    JOIN iam_user_identifiers AS recipient
      ON recipient.user_id = setup.user_id
     AND recipient.kind = 'email_v1' AND recipient.state = 'active'
    WHERE setup.binding_id = p_binding_id
      AND setup.setup_state = 'required' AND binding.state = 'pending';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM portal setup target is unavailable';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION iam_email_verification_delivery_target_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_password_reset_delivery_target_v1(
    name,text,bytea
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_portal_setup_delivery_target_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;
-- +goose StatementEnd
