-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_operational_materialize_v1(
    p_scope_schema name, p_component_id uuid, p_document jsonb,
    p_semantic_sha256 bytea, p_occurred_at timestamptz
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource_slug text;
    v_resource_kind text;
    v_resource_name text;
    v_resource_description text;
    v_grant_resource_slug text;
    v_grant_role_name text;
    v_grant_role_description text;
    v_grant_actions jsonb;
    v_action_key text;
    v_resource_id uuid;
    v_role_id uuid;
    v_action_id uuid;
    v_permission_id uuid;
    v_binding iam_role_permission_bindings%ROWTYPE;
    v_clock bigint;
    v_count bigint;
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM iam_catalog_materializations AS materialization
        WHERE materialization.component_id=p_component_id
          AND materialization.semantic_sha256=p_semantic_sha256
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog projection prerequisite differs';
    END IF;
    FOR v_resource_slug,v_resource_kind,
        v_resource_name,v_resource_description IN
        SELECT resource.slug,resource.kind,resource.name,resource.description
        FROM pg_catalog.jsonb_to_recordset(p_document->'resources')
            AS resource(slug text,kind text,name text,description text)
    LOOP
        v_resource_id := iam_catalog_operational_resource_v1(
            p_component_id,v_resource_slug,v_resource_kind,
            v_resource_name,v_resource_description,p_occurred_at
        );
    END LOOP;
    FOR v_grant_resource_slug,v_grant_role_name,
        v_grant_role_description,v_grant_actions IN
        SELECT grant_row.resource_slug,grant_row.role_name,
               grant_row.role_description,grant_row.actions
        FROM pg_catalog.jsonb_to_recordset(p_document->'grants')
            AS grant_row(
                resource_slug text,role_name text,
                role_description text,actions jsonb
            )
    LOOP
        SELECT resource.resource_id INTO STRICT v_resource_id
        FROM iam_catalog_operational_resources AS resource
        WHERE resource.component_id=p_component_id
          AND resource.resource_slug=v_grant_resource_slug;
        v_role_id := iam_catalog_operational_role_v1(
            p_scope_schema,v_grant_role_name,
            v_grant_role_description,p_occurred_at
        );
        FOR v_action_key IN
            SELECT pg_catalog.jsonb_array_elements_text(v_grant_actions)
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
            ELSIF v_binding.state IS DISTINCT FROM 'active' THEN
                RAISE EXCEPTION USING ERRCODE='23505',
                    MESSAGE='IAM catalog operational binding differs';
            END IF;
            INSERT INTO iam_catalog_operational_grants(
                component_id,resource_slug,role_name,action_key,
                role_id,resource_id,action_id,permission_id,binding_id
            ) VALUES (
                p_component_id,v_grant_resource_slug,v_grant_role_name,
                v_action_key,v_role_id,v_resource_id,v_action_id,
                v_permission_id,v_binding.id
            );
        END LOOP;
    END LOOP;
    UPDATE iam_authorization_clock
    SET clock_value=clock_value+1,revision=revision+1,
        updated_at=GREATEST(updated_at,p_occurred_at)
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM authorization clock is unavailable';
    END IF;
    v_count := iam_rbac_bump_all_users_v1(v_clock);
    INSERT INTO iam_authorization_operation_receipts(
        operation_id,command_key,request_sha256,clock_value,
        result_code,result_id,result_revision,result_sha256,created_at
    ) VALUES (
        p_component_id,'iam.catalog.materialize',p_semantic_sha256,
        v_clock,'iam.catalog.materialized',p_component_id,1,
        p_semantic_sha256,p_occurred_at
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_catalog_operational_materialize_v1(
    name,uuid,jsonb,bytea,timestamptz
) FROM PUBLIC;
