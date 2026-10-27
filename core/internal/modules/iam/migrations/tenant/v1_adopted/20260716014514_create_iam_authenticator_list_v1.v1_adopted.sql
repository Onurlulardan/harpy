-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_authenticator_list_v1(
    p_scope_schema name, p_user_id uuid, p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint,
    p_limit integer, p_after_updated_at timestamptz, p_after_id uuid
)
RETURNS TABLE (
    authenticator_id uuid, user_id uuid, kind text, display_label text,
    state text, revision bigint, created_at timestamptz,
    updated_at timestamptz, activated_at timestamptz,
    last_used_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_user_id IS NULL OR p_limit NOT BETWEEN 1 AND 100 OR
       ((p_after_updated_at IS NULL) <> (p_after_id IS NULL)) OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501';
    END IF;
    RETURN QUERY
    SELECT authenticator.id, authenticator.user_id, authenticator.kind,
        authenticator.display_label, authenticator.state,
        authenticator.revision, authenticator.created_at,
        authenticator.updated_at, authenticator.activated_at,
        authenticator.last_used_at
    FROM iam_authenticators AS authenticator
    WHERE authenticator.user_id = p_user_id
      AND authenticator.state <> 'retired'
      AND (p_after_updated_at IS NULL OR
          (authenticator.updated_at, authenticator.id) <
          (p_after_updated_at, p_after_id))
    ORDER BY authenticator.updated_at DESC, authenticator.id DESC
    LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_authenticator_list_v1(
    name,uuid,uuid,uuid,bigint,bigint,integer,timestamptz,uuid
) FROM PUBLIC;
