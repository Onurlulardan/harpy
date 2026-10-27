-- +goose Up
-- Migration stream version: 20260728105100.
-- +goose StatementBegin
CREATE FUNCTION catalog_retract_empty_menu_from_v1_target_v1(
    p_expected_owner name
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_origin text;
    v_tenant_id uuid;
    v_group_state text;
    v_group_writer text;
    v_group_generation bigint;
    v_domain_state text;
    v_domain_writer text;
    v_domain_generation bigint;
    v_member_count bigint;
    v_existing bigint;
BEGIN
    IF p_expected_owner IS DISTINCT FROM 'forgecrud_owner'::name
       OR CURRENT_USER IS DISTINCT FROM p_expected_owner THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Catalog V1 target repair owner denied';
    END IF;
    -- The schema identity sentinel is written by the final tenant migration, so
    -- a first-run fresh schema has no row here and carries no V1 target.
    SELECT identity.origin, identity.tenant_id
    INTO v_origin, v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton;
    IF NOT FOUND THEN
        RETURN;
    END IF;

    IF v_origin = 'v1_adopted' THEN
        SELECT authority.state, authority.writer, authority.generation
        INTO v_group_state, v_group_writer, v_group_generation
        FROM platform.cutover_groups AS authority
        WHERE authority.scope_kind = 'tenant'
          AND authority.scope_id = v_tenant_id
          AND authority.group_id = 'catalog'
        FOR SHARE;
        IF NOT FOUND OR v_group_generation < 1 THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'Catalog singleton cutover group differs';
        END IF;

        SELECT pg_catalog.count(*)
        INTO v_member_count
        FROM platform.cutover_group_members AS member
        WHERE member.scope_kind = 'tenant'
          AND member.scope_id = v_tenant_id
          AND member.group_id = 'catalog'
          AND member.ordinal = 1
          AND member.domain = 'catalog';
        IF v_member_count <> 1 OR (
            SELECT pg_catalog.count(*)
            FROM platform.cutover_group_members AS member
            WHERE member.scope_kind = 'tenant'
              AND member.scope_id = v_tenant_id
              AND member.group_id = 'catalog'
        ) <> 1 THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'Catalog singleton cutover member differs';
        END IF;

        SELECT domain.state, domain.writer, domain.generation
        INTO v_domain_state, v_domain_writer, v_domain_generation
        FROM platform.domain_cutovers AS domain
        WHERE domain.scope_kind = 'tenant'
          AND domain.scope_id = v_tenant_id
          AND domain.domain = 'catalog'
        FOR SHARE;
        IF NOT FOUND OR v_domain_state IS DISTINCT FROM v_group_state
           OR v_domain_writer IS DISTINCT FROM v_group_writer
           OR v_domain_generation IS DISTINCT FROM v_group_generation THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'Catalog singleton cutover authority differs';
        END IF;

        IF v_group_state = 'v2' AND v_group_writer = 'v2' THEN
            NULL;
        ELSIF (v_group_state IN ('v1', 'shadow') AND v_group_writer = 'v1')
           OR (v_group_state IN ('quiescing', 'fenced')
               AND v_group_writer = 'none') THEN
            SELECT
              (SELECT pg_catalog.count(*) FROM catalog_installations)+
              (SELECT pg_catalog.count(*) FROM catalog_component_bindings)+
              (SELECT pg_catalog.count(*) FROM catalog_dependency_resolutions)+
              (SELECT pg_catalog.count(*) FROM catalog_operation_runs)+
              (SELECT pg_catalog.count(*) FROM catalog_operation_steps)+
              (SELECT pg_catalog.count(*) FROM catalog_command_receipts)+
              (SELECT pg_catalog.count(*) FROM catalog_purge_blocker_receipts)+
              (SELECT pg_catalog.count(*) FROM catalog_demo_receipts)+
              (SELECT pg_catalog.count(*) FROM catalog_menu_definitions)+
              (SELECT pg_catalog.count(*) FROM catalog_menu_versions)+
              (SELECT pg_catalog.count(*) FROM catalog_menu_items)+
              (SELECT pg_catalog.count(*) FROM catalog_menu_overrides)+
              (SELECT pg_catalog.count(*) FROM catalog_menu_override_receipts)+
              (SELECT pg_catalog.count(*) FROM catalog_catalog_components)+
              (SELECT pg_catalog.count(*) FROM catalog_catalog_component_receipts)+
              (SELECT pg_catalog.count(*) FROM catalog_catalog_component_blockers)+
              (SELECT pg_catalog.count(*) FROM catalog_catalog_blocker_receipts)
            INTO v_existing;
            IF v_existing <> 2 OR NOT EXISTS (
                SELECT 1
                FROM catalog_menu_definitions AS definition
                JOIN catalog_menu_versions AS version
                  ON version.version_id = definition.active_version_id
                 AND version.definition_id = definition.definition_id
                WHERE definition.definition_id =
                      '018f6000-0000-7000-8000-000000000001'::uuid
                  AND definition.revision = 1
                  AND definition.active_version_id =
                      '018f6000-0000-7000-8000-000000000002'::uuid
                  AND version.version_number = 1
                  AND version.content_sha256 =
                      'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
                  AND NOT EXISTS (
                      SELECT 1 FROM catalog_menu_items AS item
                      WHERE item.version_id = version.version_id
                  )
            ) THEN
                RAISE EXCEPTION USING ERRCODE = '55000',
                    MESSAGE =
                      'Catalog V1 inactive target is not the exact baseline';
            END IF;

            UPDATE catalog_menu_definitions
            SET active_version_id = NULL,
                updated_at = pg_catalog.clock_timestamp()
            WHERE definition_id =
                  '018f6000-0000-7000-8000-000000000001'::uuid
              AND revision = 1
              AND active_version_id =
                  '018f6000-0000-7000-8000-000000000002'::uuid;
            IF NOT FOUND THEN
                RAISE EXCEPTION USING ERRCODE = '55000',
                    MESSAGE = 'Catalog V1 empty-menu baseline changed';
            END IF;

            DELETE FROM catalog_menu_versions
            WHERE version_id =
                  '018f6000-0000-7000-8000-000000000002'::uuid
              AND definition_id =
                  '018f6000-0000-7000-8000-000000000001'::uuid
              AND version_number = 1
              AND content_sha256 =
                  'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855';
            IF NOT FOUND THEN
                RAISE EXCEPTION USING ERRCODE = '55000',
                    MESSAGE = 'Catalog V1 empty-menu version changed';
            END IF;

            DELETE FROM catalog_menu_definitions
            WHERE definition_id =
                  '018f6000-0000-7000-8000-000000000001'::uuid
              AND revision = 1
              AND active_version_id IS NULL;
            IF NOT FOUND THEN
                RAISE EXCEPTION USING ERRCODE = '55000',
                    MESSAGE = 'Catalog V1 empty-menu definition changed';
            END IF;
        ELSE
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'Catalog singleton cutover authority state invalid';
        END IF;
    END IF;
    RETURN;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_retract_empty_menu_from_v1_target_v1(name)
FROM PUBLIC;

SELECT catalog_retract_empty_menu_from_v1_target_v1('forgecrud_owner'::name);

DROP FUNCTION catalog_retract_empty_menu_from_v1_target_v1(name);
