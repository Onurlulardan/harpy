-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_grants_set_v2(
    p_component_id uuid,p_enable boolean,p_completed_at timestamptz
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_binding_id uuid;
    v_changed bigint:=0;
    v_step bigint;
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_enable IS NULL OR p_completed_at IS NULL OR
       NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at OR
       NOT EXISTS (
           SELECT 1 FROM iam_catalog_component_heads AS head
           WHERE head.component_id=p_component_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM Catalog grant lifecycle input is invalid';
    END IF;
    FOR v_binding_id IN
        SELECT DISTINCT mapped.binding_id
        FROM iam_catalog_version_operational_grants AS mapped
        JOIN iam_catalog_component_heads AS head
          ON head.component_id=mapped.component_id
         AND head.package_id=mapped.package_id
         AND head.package_version=mapped.package_version
         AND head.release_id=mapped.release_id
         AND head.owner_generation=mapped.owner_generation
         AND head.content_sha256=mapped.content_sha256
        WHERE mapped.component_id=p_component_id
        ORDER BY mapped.binding_id
    LOOP
        IF p_enable THEN
            UPDATE iam_role_permission_bindings AS binding
            SET state='active',revision=binding.revision+1,
                updated_at=p_completed_at,activated_at=p_completed_at,
                revoked_at=NULL
            WHERE binding.id=v_binding_id AND binding.state='revoked';
        ELSE
            UPDATE iam_role_permission_bindings AS binding
            SET state='revoked',revision=binding.revision+1,
                updated_at=p_completed_at,revoked_at=p_completed_at
            WHERE binding.id=v_binding_id AND binding.state='active'
              AND NOT EXISTS (
                  SELECT 1
                  FROM iam_catalog_version_operational_grants AS other
                  JOIN iam_catalog_component_heads AS other_head
                    ON other_head.component_id=other.component_id
                   AND other_head.package_id=other.package_id
                   AND other_head.package_version=other.package_version
                   AND other_head.release_id=other.release_id
                   AND other_head.owner_generation=other.owner_generation
                   AND other_head.content_sha256=other.content_sha256
                  JOIN iam_catalog_components AS other_component
                    ON other_component.component_id=other.component_id
                  WHERE other.binding_id=v_binding_id
                    AND other.component_id<>p_component_id
                    AND other_component.lifecycle_state='ACTIVE'
              );
        END IF;
        GET DIAGNOSTICS v_step=ROW_COUNT;
        v_changed:=v_changed+v_step;
    END LOOP;
    IF p_enable AND EXISTS (
        SELECT 1
        FROM iam_catalog_version_operational_grants AS mapped
        JOIN iam_catalog_component_heads AS head
          ON head.component_id=mapped.component_id
         AND head.package_id=mapped.package_id
         AND head.package_version=mapped.package_version
         AND head.release_id=mapped.release_id
         AND head.owner_generation=mapped.owner_generation
         AND head.content_sha256=mapped.content_sha256
        JOIN iam_role_permission_bindings AS binding
          ON binding.id=mapped.binding_id
        WHERE mapped.component_id=p_component_id
          AND binding.state<>'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog grants could not be activated';
    END IF;
    RETURN v_changed;
END;
$function$;

CREATE FUNCTION iam_catalog_authorization_record_v2(
    p_operation_id uuid,p_component_id uuid,p_content_sha256 text,
    p_action text,p_completed_at timestamptz
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_clock bigint;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_action NOT IN ('disable','enable','retire','restore','purge') OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM Catalog authorization receipt input is invalid';
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value=clock_value+1,revision=revision+1,
        updated_at=GREATEST(updated_at,p_completed_at)
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM Catalog authorization clock is unavailable';
    END IF;
    PERFORM iam_rbac_bump_all_users_v1(v_clock);
    INSERT INTO iam_authorization_operation_receipts(
        operation_id,command_key,request_sha256,clock_value,
        result_code,result_id,result_revision,result_sha256,created_at
    ) VALUES (
        p_operation_id,'iam.catalog.'||p_action,
        pg_catalog.decode(p_content_sha256,'hex'),v_clock,
        'iam.catalog.'||p_action||'.succeeded',p_component_id,1,
        pg_catalog.decode(p_content_sha256,'hex'),p_completed_at
    );
    RETURN v_clock;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_grants_set_v2(
    uuid,boolean,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_catalog_authorization_record_v2(
    uuid,uuid,text,text,timestamptz
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_catalog_grants_set_v2(
    uuid,boolean,timestamptz
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION iam_catalog_authorization_record_v2(
    uuid,uuid,text,text,timestamptz
) FROM forgecrud_runtime;
