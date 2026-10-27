-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_assert_current_authority_v1(p_scope_schema name)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_authority_id uuid;
    v_schema text;
    v_tenant_id uuid;
BEGIN
    v_schema := pg_catalog.current_schema()::text;
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS NULL OR p_scope_schema::text IS DISTINCT FROM v_schema THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics runtime authority role is invalid';
    END IF;
    BEGIN
        SELECT identity_row.tenant_id INTO STRICT v_tenant_id
        FROM tenancy_schema_identity AS identity_row
        WHERE identity_row.singleton AND identity_row.schema_name = v_schema;
    EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'analytics current TenantScope identity is invalid';
    END;
    SELECT authority.authority_id INTO STRICT v_authority_id
    FROM platform.analytics_resolve_current_tenant_authority_v1(v_tenant_id) AS authority;
    RETURN v_authority_id;
END;
$function$;

CREATE FUNCTION analytics_reject_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'analytics immutable row cannot be changed';
END;
$function$;

CREATE FUNCTION analytics_guard_operation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'analytics operation receipt is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION analytics_assert_current_authority_v1(name) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_reject_mutation_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_guard_operation_v1() FROM PUBLIC;

CREATE TRIGGER analytics_query_dependencies_immutable
BEFORE UPDATE OR DELETE ON analytics_query_dependencies
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_v1();
CREATE TRIGGER analytics_publication_approvals_immutable
BEFORE UPDATE OR DELETE ON analytics_publication_approvals
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_v1();
CREATE TRIGGER analytics_execution_audits_immutable
BEFORE UPDATE OR DELETE ON analytics_execution_audits
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_v1();
CREATE TRIGGER analytics_adoption_acceptance_immutable
BEFORE UPDATE OR DELETE ON analytics_v1_adoption_acceptance_receipts
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_v1();
CREATE TRIGGER analytics_adoption_items_immutable
BEFORE UPDATE OR DELETE ON analytics_v1_adoption_items
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_v1();
CREATE TRIGGER analytics_operations_immutable
BEFORE INSERT OR UPDATE OR DELETE ON analytics_operations
FOR EACH ROW EXECUTE FUNCTION analytics_guard_operation_v1();

ALTER TABLE analytics_query_dependencies
    ENABLE ALWAYS TRIGGER analytics_query_dependencies_immutable;
ALTER TABLE analytics_publication_approvals
    ENABLE ALWAYS TRIGGER analytics_publication_approvals_immutable;
ALTER TABLE analytics_execution_audits
    ENABLE ALWAYS TRIGGER analytics_execution_audits_immutable;
ALTER TABLE analytics_v1_adoption_acceptance_receipts
    ENABLE ALWAYS TRIGGER analytics_adoption_acceptance_immutable;
ALTER TABLE analytics_v1_adoption_items
    ENABLE ALWAYS TRIGGER analytics_adoption_items_immutable;
ALTER TABLE analytics_operations
    ENABLE ALWAYS TRIGGER analytics_operations_immutable;
-- +goose StatementEnd
