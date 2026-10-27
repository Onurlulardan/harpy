-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_portal_authorization_evidence_read_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,
    p_resource_key text,p_action_key text
)
RETURNS TABLE (
    actor_document jsonb,resource_document jsonb,action_document jsonb,
    authorization_clock bigint,bindings_document jsonb,implications_document jsonb
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_actor_id IS NULL OR p_resource_key IS NULL OR p_action_key IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501';
    END IF;
    RETURN QUERY
    SELECT evidence.actor_document,evidence.resource_document,evidence.action_document,
      evidence.authorization_clock,
      COALESCE((
        SELECT pg_catalog.jsonb_agg(binding_element ORDER BY binding_element->>'binding_id')
        FROM pg_catalog.jsonb_array_elements(evidence.bindings_document) binding_element
        WHERE binding_element->>'path'='role' AND EXISTS(
          SELECT 1 FROM iam_portal_role_assignments portal_role
          JOIN iam_portal_subject_bindings portal_binding
            ON portal_binding.id=portal_role.binding_id
            AND portal_binding.user_id=portal_role.user_id
          JOIN iam_user_access_modes portal_access
            ON portal_access.id=portal_binding.access_mode_id
            AND portal_access.user_id=portal_binding.user_id
          JOIN iam_portal_access_setups portal_setup
            ON portal_setup.binding_id=portal_binding.id
          WHERE portal_role.user_id=p_actor_id
            AND portal_role.assignment_id=(binding_element->>'assignment_id')::uuid
            AND portal_binding.state='active' AND portal_access.state='active'
            AND portal_setup.setup_state='completed'
            AND portal_setup.approval_state IN('not_required','approved')
        )
      ),'[]'::jsonb),
      evidence.implications_document
    FROM iam_authorization_evidence_read_v3(
      p_scope_schema,p_tenant_id,p_actor_id,p_resource_key,p_action_key
    ) evidence;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_authorization_evidence_read_v1(
    name,uuid,uuid,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_portal_authorization_evidence_read_v1(
    name,uuid,uuid,text,text
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_user_create_v1(
    name,uuid,uuid,bytea,uuid,uuid,bytea,bytea,uuid,text,text,bytea,
    uuid,text,text,bytea,uuid,uuid,text,uuid,uuid,uuid,uuid,bytea,
    text,text[],uuid[],uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_user_status_v1(
    name,uuid,uuid,bytea,uuid,bigint,text,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_user_console_grant_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_user_console_revoke_v1(
    name,uuid,uuid,bytea,uuid,bigint,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_user_read_v1(name,uuid)
    TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_access_verify_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_subject_links_read_v1(
    name,uuid,uuid,bigint,bigint,bigint,uuid,uuid,uuid[]
) TO harpy_runtime;
