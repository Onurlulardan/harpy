-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_project_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_run_id uuid
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_target_table text;
    v_target_record jsonb;
    v_handled boolean;
    v_projected bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    v_projected := 0;
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT EXISTS (
           SELECT 1 FROM iam_v1_adoption_runs AS run
           WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
             AND run.eligible AND run.state = 'copying'
           FOR UPDATE
       ) OR EXISTS (
           SELECT 1 FROM iam_v1_adoption_category_evidence AS evidence
           WHERE evidence.operation_id = p_operation_id AND evidence.run_id = p_run_id
             AND evidence.category <> 'layouts' AND evidence.copy_state <> 'complete'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 adoption projection authority or readiness denied';
    END IF;
    FOR v_target_table, v_target_record IN
        SELECT stage.target_table, stage.target_record
        FROM iam_v1_adoption_stage AS stage
        WHERE stage.operation_id = p_operation_id AND stage.run_id = p_run_id
          AND NOT stage.removed
        ORDER BY CASE stage.target_table
            WHEN 'iam_profile_schema_versions' THEN 10
            WHEN 'iam_profile_fields' THEN 20
            WHEN 'iam_profile_field_revisions' THEN 30
            WHEN 'iam_organizations' THEN 40
            WHEN 'iam_users' THEN 50
            WHEN 'iam_user_identifiers' THEN 60
            WHEN 'iam_password_credentials' THEN 70
            WHEN 'iam_organization_memberships' THEN 80
            WHEN 'iam_roles' THEN 90
            WHEN 'iam_user_role_assignments' THEN 100
            WHEN 'iam_user_access_modes' THEN 110
            WHEN 'iam_portal_subject_bindings' THEN 120
            WHEN 'iam_portal_access_setups' THEN 130
            WHEN 'iam_resources' THEN 140
            WHEN 'iam_actions' THEN 150
            WHEN 'iam_permissions' THEN 160
            WHEN 'iam_user_permission_bindings' THEN 170
            WHEN 'iam_role_permission_bindings' THEN 180
            WHEN 'iam_organization_permission_bindings' THEN 190
            WHEN 'iam_authorization_clock' THEN 200
            WHEN 'iam_protected_admin_policies' THEN 210
            WHEN 'iam_permission_request_groups' THEN 220
            WHEN 'iam_permission_requests' THEN 230
            WHEN 'iam_invitations' THEN 240
            WHEN 'iam_one_time_tokens' THEN 250
        END, stage.target_key
    LOOP
        v_handled := iam_v1_adoption_project_identity_record_v1(
            v_target_table, v_target_record
        );
        IF NOT v_handled THEN
            v_handled := iam_v1_adoption_project_access_record_v1(
                v_target_table, v_target_record
            );
        END IF;
        IF NOT v_handled THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'IAM V1 adoption projection target is unsupported';
        END IF;
        v_projected := v_projected + 1;
    END LOOP;
    UPDATE iam_v1_adoption_runs
    SET state = 'projected', projected_at = v_now, updated_at = v_now
    WHERE operation_id = p_operation_id AND run_id = p_run_id AND state = 'copying';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '22000',
            MESSAGE = 'IAM V1 adoption projection state raced';
    END IF;
    RETURN v_projected;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_project_v1(
    name, uuid, uuid
) FROM PUBLIC;
