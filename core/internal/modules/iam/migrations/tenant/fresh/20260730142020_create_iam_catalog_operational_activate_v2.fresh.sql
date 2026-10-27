-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_operational_activate_v2(
    p_scope_schema name,p_provision_receipt_id uuid,
    p_occurred_at timestamptz
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_provision iam_catalog_version_provisions%ROWTYPE;
    v_resource iam_catalog_version_resources%ROWTYPE;
    v_grant iam_catalog_version_grants%ROWTYPE;
    v_action_key text;
    v_resource_id uuid;
    v_role_id uuid;
    v_action_id uuid;
    v_permission_id uuid;
    v_binding iam_role_permission_bindings%ROWTYPE;
    v_clock bigint;
BEGIN
    SELECT * INTO v_provision
    FROM iam_catalog_version_provisions AS provision
    WHERE provision.receipt_id=p_provision_receipt_id
    FOR SHARE;
    IF NOT FOUND OR p_scope_schema IS DISTINCT FROM (
           SELECT identity_row.schema_name::name
           FROM tenancy_schema_identity AS identity_row
           WHERE identity_row.singleton
       ) OR p_occurred_at IS NULL OR
       NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM catalog version activation input is invalid';
    END IF;
    FOR v_resource IN
        SELECT resource.*
        FROM iam_catalog_version_resources AS resource
        WHERE resource.component_id=v_provision.component_id
          AND resource.package_id=v_provision.package_id
          AND resource.package_version=v_provision.package_version
          AND resource.release_id=v_provision.release_id
          AND resource.owner_generation=v_provision.owner_generation
          AND resource.content_sha256=v_provision.content_sha256
        ORDER BY resource.resource_slug
    LOOP
        v_resource_id := iam_catalog_operational_resource_v2(
            v_provision.component_id,v_provision.package_id,
            v_provision.package_version,v_provision.release_id,
            v_provision.owner_generation,v_provision.content_sha256,
            v_resource.resource_slug,v_resource.resource_kind,
            v_resource.name,v_resource.description,p_occurred_at
        );
    END LOOP;
    FOR v_grant IN
        SELECT grant_row.*
        FROM iam_catalog_version_grants AS grant_row
        WHERE grant_row.component_id=v_provision.component_id
          AND grant_row.package_id=v_provision.package_id
          AND grant_row.package_version=v_provision.package_version
          AND grant_row.release_id=v_provision.release_id
          AND grant_row.owner_generation=v_provision.owner_generation
          AND grant_row.content_sha256=v_provision.content_sha256
        ORDER BY grant_row.resource_slug,grant_row.role_name
    LOOP
        SELECT projected.resource_id INTO STRICT v_resource_id
        FROM iam_catalog_version_operational_resources AS projected
        WHERE projected.component_id=v_provision.component_id
          AND projected.package_id=v_provision.package_id
          AND projected.package_version=v_provision.package_version
          AND projected.release_id=v_provision.release_id
          AND projected.owner_generation=v_provision.owner_generation
          AND projected.content_sha256=v_provision.content_sha256
          AND projected.resource_slug=v_grant.resource_slug;
        v_role_id := iam_catalog_operational_role_v2(
            p_scope_schema,v_grant.role_name,
            v_grant.role_description,p_occurred_at
        );
        FOREACH v_action_key IN ARRAY v_grant.actions
        LOOP
            v_action_id := iam_catalog_operational_action_v1(
                v_action_key,p_occurred_at
            );
            v_permission_id := iam_catalog_operational_permission_v1(
                v_resource_id,v_action_id,p_occurred_at
            );
            SELECT * INTO v_binding
            FROM iam_role_permission_bindings AS binding
            WHERE binding.role_id=v_role_id
              AND binding.permission_id=v_permission_id
            FOR UPDATE;
            IF NOT FOUND THEN
                INSERT INTO iam_role_permission_bindings(
                    id,role_id,role_is_protected,permission_id,
                    permission_match_kind,state,revision,created_at,
                    updated_at,activated_at
                ) VALUES (
                    pg_catalog.uuidv7(),v_role_id,false,v_permission_id,
                    'exact','active',1,p_occurred_at,
                    p_occurred_at,p_occurred_at
                ) RETURNING * INTO v_binding;
            ELSIF v_binding.state='revoked' THEN
                UPDATE iam_role_permission_bindings
                SET state='active',revision=revision+1,
                    updated_at=p_occurred_at,activated_at=p_occurred_at,
                    revoked_at=NULL
                WHERE id=v_binding.id
                RETURNING * INTO v_binding;
            ELSIF v_binding.state<>'active' THEN
                RAISE EXCEPTION USING ERRCODE='23505',
                    MESSAGE='IAM catalog version binding differs';
            END IF;
            INSERT INTO iam_catalog_version_operational_grants(
                component_id,package_id,package_version,release_id,
                owner_generation,content_sha256,resource_slug,
                role_name,action_key,role_id,resource_id,action_id,
                permission_id,binding_id
            ) VALUES (
                v_provision.component_id,v_provision.package_id,
                v_provision.package_version,v_provision.release_id,
                v_provision.owner_generation,v_provision.content_sha256,
                v_grant.resource_slug,v_grant.role_name,v_action_key,
                v_role_id,v_resource_id,v_action_id,
                v_permission_id,v_binding.id
            ) ON CONFLICT DO NOTHING;
            IF NOT EXISTS (
                SELECT 1
                FROM iam_catalog_version_operational_grants AS projected
                WHERE projected.component_id=v_provision.component_id
                  AND projected.package_id=v_provision.package_id
                  AND projected.package_version=v_provision.package_version
                  AND projected.release_id=v_provision.release_id
                  AND projected.owner_generation=v_provision.owner_generation
                  AND projected.content_sha256=v_provision.content_sha256
                  AND projected.resource_slug=v_grant.resource_slug
                  AND projected.role_name=v_grant.role_name
                  AND projected.action_key=v_action_key
                  AND projected.role_id=v_role_id
                  AND projected.resource_id=v_resource_id
                  AND projected.action_id=v_action_id
                  AND projected.permission_id=v_permission_id
                  AND projected.binding_id=v_binding.id
            ) THEN
                RAISE EXCEPTION USING ERRCODE='23505',
                    MESSAGE='IAM catalog version grant mapping differs';
            END IF;
        END LOOP;
    END LOOP;
    IF NOT iam_catalog_operational_readback_v2(
        v_provision.component_id,v_provision.package_id,
        v_provision.package_version,v_provision.release_id,
        v_provision.owner_generation,v_provision.content_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog version operational readback differs';
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value=clock_value+1,revision=revision+1,
        updated_at=GREATEST(updated_at,p_occurred_at)
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM authorization clock is unavailable';
    END IF;
    PERFORM iam_rbac_bump_all_users_v1(v_clock);
    INSERT INTO iam_authorization_operation_receipts(
        operation_id,command_key,request_sha256,clock_value,
        result_code,result_id,result_revision,result_sha256,created_at
    ) VALUES (
        v_provision.receipt_id,'iam.catalog.activate',
        v_provision.semantic_sha256,v_clock,'iam.catalog.activated',
        v_provision.component_id,1,
        v_provision.semantic_sha256,p_occurred_at
    );
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_operational_activate_v2(
    name,uuid,timestamptz
) FROM PUBLIC;
