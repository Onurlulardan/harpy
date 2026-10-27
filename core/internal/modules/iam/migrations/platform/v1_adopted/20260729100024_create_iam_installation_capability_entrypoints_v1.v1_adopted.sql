-- +goose Up
-- Migration stream version: 20260729100024.
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_activation_assert_v2(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_schema_name name
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN SESSION_USER = 'harpy_v2_runtime'
       AND CURRENT_USER = 'harpy_owner'
       AND (
         platform.tenancy_activation_step_context_assert_v1(
            p_operation_id, 'iam.initialize_tenant_identity',
            p_environment_id, p_tenant_id, p_schema_name,
            'environment_initialize'
         )
         OR platform.tenancy_activation_step_context_assert_v1(
            p_operation_id, 'iam.initialize_tenant_identity',
            p_environment_id, p_tenant_id, p_schema_name,
            'saas_customer_tenant_provision'
         )
       );
END;
$function$;

CREATE FUNCTION platform.iam_installation_security_epoch_initialize_v2(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_schema_name name, p_now timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_now IS NULL OR NOT pg_catalog.isfinite(p_now)
       OR NOT platform.iam_installation_activation_assert_v2(
            p_operation_id, p_environment_id, p_tenant_id, p_schema_name
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM installation security epoch capability required';
    END IF;
    INSERT INTO platform.iam_security_epochs (
        environment_id, security_epoch, generation, created_at, rotated_at
    ) VALUES (
        p_environment_id,
        pg_catalog.sha256(
            pg_catalog.uuid_send(pg_catalog.gen_random_uuid()) ||
            pg_catalog.uuid_send(pg_catalog.gen_random_uuid())
        ),
        1, p_now, p_now
    ) ON CONFLICT (environment_id) DO NOTHING;
    RETURN true;
END;
$function$;

CREATE FUNCTION platform.iam_installation_identity_directory_upsert_v1(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_schema_name name, p_group_generation bigint, p_iam_generation bigint,
    p_user_id uuid, p_identifier_id uuid, p_lookup_key_id text,
    p_identifier_hmac bytea, p_source_revision bigint, p_source_sha256 bytea
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT platform.iam_installation_activation_assert_v2(
        p_operation_id, p_environment_id, p_tenant_id, p_schema_name
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM installation directory capability required';
    END IF;
    RETURN platform.iam_identity_directory_upsert_v2(
        p_tenant_id, p_group_generation, p_iam_generation,
        p_user_id, p_identifier_id, p_lookup_key_id,
        p_identifier_hmac, p_source_revision, p_source_sha256
    );
END;
$function$;

CREATE FUNCTION platform.iam_installation_instance_owner_bind_v2(
    p_operation_id uuid,p_environment_id uuid,p_tenant_id uuid,p_schema_name name,
    p_user_id uuid,p_subject_id uuid,p_owner_role_id uuid,
    p_role_permission_id uuid,p_owner_binding_id uuid,
    p_authorization_version bigint,p_request_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
SET application_name TO 'forgecrud-v2:core-operator-initialize'
AS $function$
DECLARE
    v_replayed boolean;
BEGIN
    IF NOT platform.tenancy_activation_step_context_assert_v1(
        p_operation_id, 'platformauthority.bind_owner', p_environment_id,
        p_tenant_id, p_schema_name, 'environment_initialize'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM installation Instance Owner capability required';
    END IF;
    v_replayed := platform.iam_installation_instance_owner_bind_v1(
        p_operation_id,p_environment_id,p_tenant_id,p_schema_name,
        p_user_id,p_subject_id,p_owner_role_id,p_role_permission_id,
        p_owner_binding_id,p_authorization_version,p_request_sha256
    );
    IF NOT platform.tenancy_activation_step_consume_v1(
        p_operation_id, 'platformauthority.bind_owner'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM installation Instance Owner capability conflicts';
    END IF;
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_activation_assert_v2(
    uuid,uuid,uuid,name
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_installation_activation_assert_v2(
    uuid,uuid,uuid,name
) TO harpy_runtime;
REVOKE ALL ON FUNCTION
    platform.iam_installation_security_epoch_initialize_v2(
        uuid,uuid,uuid,name,timestamptz
    ) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.iam_installation_security_epoch_initialize_v2(
        uuid,uuid,uuid,name,timestamptz
    ) TO harpy_runtime;
REVOKE ALL ON FUNCTION
    platform.iam_installation_identity_directory_upsert_v1(
        uuid,uuid,uuid,name,bigint,bigint,uuid,uuid,text,bytea,bigint,bytea
    ) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.iam_installation_identity_directory_upsert_v1(
        uuid,uuid,uuid,name,bigint,bigint,uuid,uuid,text,bytea,bigint,bytea
    ) TO harpy_runtime;
REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_bind_v2(
    uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_bind_v2(
    uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text
) TO harpy_runtime;
