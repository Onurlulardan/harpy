-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_assignment_step_up_purpose_v1(
    p_command_key text, p_target_id uuid
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RETURN pg_catalog.sha256(pg_catalog.convert_to(
        'harpy/iam/rbac-step-up/v1|' || p_command_key || '|' ||
        CASE WHEN p_command_key = 'iam.rbac.assignment.create'
            THEN '' ELSE coalesce(p_target_id::text, '') END,
        'UTF8'
    ));
END;
$function$;
REVOKE ALL ON FUNCTION iam_rbac_assignment_step_up_purpose_v1(text,uuid)
    FROM PUBLIC;

CREATE FUNCTION iam_rbac_assignment_step_up_assert_v1(
    p_command_key text, p_target_id uuid, p_purpose_sha256 bytea,
    p_auth_time timestamptz, p_amr_password boolean,
    p_amr_webauthn boolean, p_amr_totp boolean
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF p_command_key NOT IN (
        'iam.rbac.assignment.create',
        'iam.rbac.assignment.update',
        'iam.rbac.assignment.retire'
    ) OR (p_command_key <> 'iam.rbac.assignment.create' AND p_target_id IS NULL) OR
       p_purpose_sha256 IS DISTINCT FROM
           iam_rbac_assignment_step_up_purpose_v1(p_command_key, p_target_id) OR
       p_auth_time IS NULL OR p_auth_time > v_now OR
       p_auth_time < v_now - INTERVAL '10 minutes' OR
       NOT (p_amr_password OR p_amr_webauthn OR p_amr_totp) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC assignment step-up evidence is denied';
    END IF;
    RETURN true;
END;
$function$;
REVOKE ALL ON FUNCTION iam_rbac_assignment_step_up_assert_v1(
    text,uuid,bytea,timestamptz,boolean,boolean,boolean
) FROM PUBLIC;

CREATE FUNCTION iam_rbac_assignment_target_active_v1(
    p_target_kind text, p_target_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE v_active boolean;
BEGIN
    CASE p_target_kind
    WHEN 'USER' THEN
        SELECT true INTO v_active FROM iam_users
        WHERE id = p_target_id AND status = 'active';
    WHEN 'ROLE' THEN
        SELECT true INTO v_active FROM iam_roles
        WHERE id = p_target_id AND status = 'active';
    WHEN 'ORGANIZATION' THEN
        SELECT true INTO v_active FROM iam_organizations
        WHERE id = p_target_id AND status = 'active';
    ELSE
        v_active := false;
    END CASE;
    RETURN coalesce(v_active, false);
END;
$function$;
REVOKE ALL ON FUNCTION iam_rbac_assignment_target_active_v1(text,uuid)
    FROM PUBLIC;

CREATE FUNCTION iam_rbac_assignment_items_valid_v1(p_items jsonb)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_count integer;
    v_valid boolean;
BEGIN
    IF pg_catalog.jsonb_typeof(p_items) IS DISTINCT FROM 'array' THEN
        RETURN false;
    END IF;
    v_count := pg_catalog.jsonb_array_length(p_items);
    IF v_count NOT BETWEEN 1 AND 100 THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.bool_and(
        pg_catalog.jsonb_typeof(item.value) = 'object' AND
        item.value ?& ARRAY['ordinal','action_id','permission_id','binding_id'] AND
        (SELECT pg_catalog.count(*) = 4
         FROM pg_catalog.jsonb_object_keys(item.value))
    ) INTO v_valid
    FROM pg_catalog.jsonb_array_elements(p_items) AS item(value);
    IF v_valid IS DISTINCT FROM true THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.count(*) = v_count AND
           pg_catalog.count(DISTINCT ordinal) = v_count AND
           pg_catalog.min(ordinal) = 1 AND pg_catalog.max(ordinal) = v_count
    INTO v_valid
    FROM pg_catalog.jsonb_to_recordset(p_items) AS item(ordinal integer);
    RETURN coalesce(v_valid, false);
END;
$function$;
REVOKE ALL ON FUNCTION iam_rbac_assignment_items_valid_v1(jsonb)
    FROM PUBLIC;

CREATE FUNCTION iam_rbac_assignment_exact_permission_v1(
    p_resource_id uuid, p_action_id uuid, p_suggested_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_permission_id uuid;
    v_now timestamptz;
BEGIN
    IF p_suggested_id IS NULL OR NOT EXISTS (
        SELECT 1 FROM iam_resources
        WHERE id = p_resource_id AND state = 'active'
    ) OR NOT EXISTS (
        SELECT 1 FROM iam_actions
        WHERE id = p_action_id AND state = 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC assignment catalog is invalid';
    END IF;
    SELECT id INTO v_permission_id FROM iam_permissions
    WHERE match_kind = 'exact' AND resource_id = p_resource_id
      AND action_id = p_action_id AND state = 'active';
    IF v_permission_id IS NOT NULL THEN
        RETURN v_permission_id;
    END IF;
    IF EXISTS (
        SELECT 1 FROM iam_permissions
        WHERE match_kind = 'exact' AND resource_id = p_resource_id
          AND action_id = p_action_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC exact permission is retired';
    END IF;
    v_now := pg_catalog.transaction_timestamp();
    INSERT INTO iam_permissions (
        id, match_kind, resource_id, action_id, is_protected, state,
        revision, created_at, updated_at, retired_at
    ) VALUES (
        p_suggested_id, 'exact', p_resource_id, p_action_id, false,
        'active', 1, v_now, v_now, NULL
    );
    RETURN p_suggested_id;
END;
$function$;
REVOKE ALL ON FUNCTION iam_rbac_assignment_exact_permission_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
-- +goose StatementEnd
