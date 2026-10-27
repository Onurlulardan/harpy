-- +goose Up
-- Migration stream version: 20260729100048.
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_instance_owner_verify_v1(
    p_operation_id uuid,p_environment_id uuid,p_tenant_id uuid,
    p_initial_user_id uuid,p_initial_subject_id uuid,p_owner_role_id uuid,
    p_role_permission_id uuid,p_initial_owner_binding_id uuid,
    p_authorization_version bigint,p_request_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.current_setting('application_name',true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_tenant_id IS NULL OR p_initial_user_id IS NULL
       OR p_initial_subject_id IS NULL OR p_owner_role_id IS NULL
       OR p_role_permission_id IS NULL OR p_initial_owner_binding_id IS NULL
       OR p_authorization_version < 1
       OR p_request_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='installation Instance Owner verification denied';
    END IF;

    PERFORM 1
    FROM platform.iam_platform_mutation_receipts AS receipt
    JOIN platform.iam_platform_subjects AS initial_subject
      ON initial_subject.environment_id=receipt.environment_id
     AND initial_subject.id=receipt.subject_id
    JOIN platform.iam_platform_subject_roles AS initial_binding
      ON initial_binding.environment_id=receipt.environment_id
     AND initial_binding.id=receipt.subject_role_binding_id
    JOIN platform.iam_platform_roles AS owner_role
      ON owner_role.environment_id=receipt.environment_id
     AND owner_role.id=receipt.role_id
    JOIN platform.iam_platform_role_permissions AS owner_permission
      ON owner_permission.environment_id=receipt.environment_id
     AND owner_permission.id=receipt.role_permission_binding_id
    JOIN platform.iam_platform_permissions AS permission
      ON permission.id=receipt.permission_id
    JOIN platform.iam_platform_owner_policies AS policy
      ON policy.environment_id=receipt.environment_id
    JOIN platform.iam_platform_subject_roles AS current_binding
      ON current_binding.environment_id=policy.environment_id
     AND current_binding.id=policy.owner_binding_id
    JOIN platform.iam_platform_subjects AS current_subject
      ON current_subject.environment_id=current_binding.environment_id
     AND current_subject.id=current_binding.subject_id
    JOIN platform.iam_platform_authorization_clocks AS clock_state
      ON clock_state.environment_id=policy.environment_id
    WHERE receipt.operation_id=p_operation_id
      AND receipt.environment_id=p_environment_id
      AND receipt.mutation_kind='initial_owner'
      AND receipt.request_sha256=pg_catalog.decode(p_request_sha256,'hex')
      AND receipt.actor_user_id=p_initial_user_id
      AND receipt.target_user_id=p_initial_user_id
      AND receipt.subject_id=p_initial_subject_id
      AND receipt.role_id=p_owner_role_id
      AND receipt.subject_role_binding_id=p_initial_owner_binding_id
      AND receipt.role_permission_binding_id=p_role_permission_id
      AND receipt.target_authorization_version=p_authorization_version
      AND receipt.platform_clock=1
      AND initial_subject.primary_tenant_id=p_tenant_id
      AND initial_subject.user_id=p_initial_user_id
      AND initial_binding.subject_id=p_initial_subject_id
      AND initial_binding.role_id=p_owner_role_id
      AND owner_role.role_kind='instance_owner'
      AND owner_role.role_key='instance_owner'
      AND owner_role.state='active'
      AND owner_permission.role_id=p_owner_role_id
      AND owner_permission.permission_id=receipt.permission_id
      AND owner_permission.state='active'
      AND permission.permission_key='environment_all'
      AND permission.match_kind='environment_all'
      AND permission.deployment_mode='both'
      AND NOT permission.delegable
      AND NOT permission.step_up_required
      AND permission.owner_only
      AND permission.state='active'
      AND current_binding.role_id=p_owner_role_id
      AND current_binding.state='active'
      AND current_subject.primary_tenant_id=p_tenant_id
      AND current_subject.state='active'
      AND clock_state.authorization_clock>=receipt.platform_clock
      AND (
          (policy.owner_binding_id=p_initial_owner_binding_id
           AND initial_binding.state='active')
          OR
          (policy.owner_binding_id<>p_initial_owner_binding_id
           AND initial_binding.state='revoked')
      )
      AND NOT EXISTS (
          SELECT 1
          FROM platform.iam_platform_mutation_receipts AS duplicate_receipt
          WHERE duplicate_receipt.environment_id=p_environment_id
            AND duplicate_receipt.mutation_kind='initial_owner'
            AND duplicate_receipt.operation_id<>p_operation_id
      )
      AND NOT EXISTS (
          SELECT 1
          FROM platform.iam_platform_subject_roles AS unexpected_binding
          JOIN platform.iam_platform_roles AS unexpected_role
            ON unexpected_role.environment_id=unexpected_binding.environment_id
           AND unexpected_role.id=unexpected_binding.role_id
          WHERE unexpected_binding.environment_id=p_environment_id
            AND unexpected_binding.state='active'
            AND unexpected_role.role_kind='instance_owner'
            AND unexpected_binding.id<>policy.owner_binding_id
      );
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='installation Instance Owner authority is inconsistent';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_verify_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_verify_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text
) TO forgecrud_runtime;
