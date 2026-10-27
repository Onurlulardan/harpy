-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_v1_adoption_row_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    selected_scope_kind text;
    selected_scope_id uuid;
    selected_authority_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 adoption principal is invalid';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        selected_scope_kind := OLD.scope_kind;
        selected_scope_id := OLD.scope_id;
        selected_authority_id := OLD.authority_id;
    ELSE
        selected_scope_kind := NEW.scope_kind;
        selected_scope_id := NEW.scope_id;
        selected_authority_id := NEW.authority_id;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'UPDATE' AND (
        NEW.scope_kind IS DISTINCT FROM OLD.scope_kind
        OR NEW.scope_id IS DISTINCT FROM OLD.scope_id
        OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing V1 adoption scope is immutable';
    END IF;
    IF selected_scope_kind IS NOT DISTINCT FROM 'global' THEN
        PERFORM platform.billing_assert_global_authority_v1(
            selected_authority_id
        );
        PERFORM 1
        FROM platform.billing_global_authorities AS authority_row
        WHERE authority_row.authority_id = selected_authority_id
          AND authority_row.scope_id = selected_scope_id;
    ELSIF selected_scope_kind IS NOT DISTINCT FROM 'tenant' THEN
        PERFORM platform.billing_assert_tenant_authority_v1(
            selected_authority_id
        );
        PERFORM 1
        FROM platform.billing_tenant_authorities AS authority_row
        WHERE authority_row.authority_id = selected_authority_id
          AND authority_row.tenant_id = selected_scope_id;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 adoption scope kind is invalid';
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'billing V1 adoption authority does not match scope';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_v1_adoption_row_v1()
    FROM PUBLIC;
-- +goose StatementEnd

CREATE TRIGGER billing_v1_adoption_runs_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_v1_adoption_runs
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_v1_adoption_row_v1();
ALTER TABLE platform.billing_v1_adoption_runs
    ENABLE ALWAYS TRIGGER billing_v1_adoption_runs_authority_guard;

CREATE TRIGGER billing_v1_adoption_categories_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_v1_adoption_categories
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_v1_adoption_row_v1();
ALTER TABLE platform.billing_v1_adoption_categories
    ENABLE ALWAYS TRIGGER billing_v1_adoption_categories_authority_guard;

CREATE TRIGGER billing_v1_adoption_mappings_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_v1_adoption_mappings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_v1_adoption_row_v1();
ALTER TABLE platform.billing_v1_adoption_mappings
    ENABLE ALWAYS TRIGGER billing_v1_adoption_mappings_authority_guard;

CREATE TRIGGER billing_v1_adoption_opening_authority_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_v1_adoption_opening_evidence
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_v1_adoption_row_v1();
ALTER TABLE platform.billing_v1_adoption_opening_evidence
    ENABLE ALWAYS TRIGGER billing_v1_adoption_opening_authority_guard;

CREATE TRIGGER billing_v1_adoption_rehearsal_authority_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_v1_adoption_rollback_rehearsals
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_v1_adoption_row_v1();
ALTER TABLE platform.billing_v1_adoption_rollback_rehearsals
    ENABLE ALWAYS TRIGGER billing_v1_adoption_rehearsal_authority_guard;

CREATE TRIGGER billing_v1_adoption_mappings_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_v1_adoption_mappings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_v1_adoption_mappings
    ENABLE ALWAYS TRIGGER billing_v1_adoption_mappings_immutable_guard;

CREATE TRIGGER billing_v1_adoption_opening_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_v1_adoption_opening_evidence
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_v1_adoption_opening_evidence
    ENABLE ALWAYS TRIGGER billing_v1_adoption_opening_immutable_guard;

CREATE TRIGGER billing_v1_adoption_rehearsal_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_v1_adoption_rollback_rehearsals
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_v1_adoption_rollback_rehearsals
    ENABLE ALWAYS TRIGGER billing_v1_adoption_rehearsal_immutable_guard;

