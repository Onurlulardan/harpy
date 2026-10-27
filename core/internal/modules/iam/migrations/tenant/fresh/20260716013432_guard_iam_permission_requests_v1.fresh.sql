-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_permission_request_terminal_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR OLD.state NOT IN (
        'APPROVED', 'REJECTED', 'CANCELLED', 'EXPIRED'
    ) THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '23514',
        CONSTRAINT = TG_TABLE_NAME || '_terminal_guard',
        MESSAGE = 'Terminal IAM permission request state is immutable';
END;
$function$;
REVOKE ALL ON FUNCTION iam_guard_permission_request_terminal_v1() FROM PUBLIC;

CREATE FUNCTION iam_guard_permission_request_group_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_group_id uuid;
    v_expected smallint;
    v_actual bigint;
    v_mismatch boolean;
BEGIN
    IF TG_TABLE_NAME = 'iam_permission_request_groups' THEN
        v_group_id := CASE WHEN TG_OP = 'DELETE' THEN OLD.id ELSE NEW.id END;
    ELSE
        v_group_id := CASE
            WHEN TG_OP = 'DELETE' THEN OLD.group_id ELSE NEW.group_id END;
    END IF;
    SELECT request_group.request_count INTO v_expected
    FROM iam_permission_request_groups AS request_group
    WHERE request_group.id = v_group_id;
    IF NOT FOUND THEN
        IF TG_OP = 'DELETE' THEN
            RETURN OLD;
        END IF;
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_permission_request_groups_cardinality_guard',
            MESSAGE = 'IAM permission request group is missing';
    END IF;
    SELECT pg_catalog.count(*), pg_catalog.bool_or(
        request.state IS DISTINCT FROM request_group.state
    ) INTO v_actual, v_mismatch
    FROM iam_permission_request_groups AS request_group
    LEFT JOIN iam_permission_requests AS request ON request.group_id = request_group.id
    WHERE request_group.id = v_group_id
    GROUP BY request_group.state;
    IF v_actual IS DISTINCT FROM v_expected OR v_mismatch IS TRUE THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_permission_request_groups_cardinality_guard',
            MESSAGE = 'IAM permission request group must be complete and state-aligned';
    END IF;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION iam_guard_permission_request_group_v1() FROM PUBLIC;
-- +goose StatementEnd

CREATE TRIGGER iam_permission_request_groups_revision_guard
BEFORE UPDATE ON iam_permission_request_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_permission_request_groups
    ENABLE ALWAYS TRIGGER iam_permission_request_groups_revision_guard;

CREATE TRIGGER iam_permission_request_groups_immutable_guard
BEFORE UPDATE ON iam_permission_request_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'requester_user_id', 'request_mode',
    'creator_authority_permission_id', 'reviewer_authority_permission_id',
    'step_up_purpose_sha256', 'request_count', 'expires_at', 'created_at'
);
ALTER TABLE iam_permission_request_groups
    ENABLE ALWAYS TRIGGER iam_permission_request_groups_immutable_guard;

CREATE TRIGGER iam_permission_request_groups_state_guard
BEFORE UPDATE ON iam_permission_request_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'PENDING>IN_REVIEW', 'PENDING>CANCELLED', 'PENDING>EXPIRED',
    'IN_REVIEW>APPROVED', 'IN_REVIEW>REJECTED',
    'IN_REVIEW>CANCELLED', 'IN_REVIEW>EXPIRED'
);
ALTER TABLE iam_permission_request_groups
    ENABLE ALWAYS TRIGGER iam_permission_request_groups_state_guard;

CREATE TRIGGER iam_permission_request_groups_terminal_guard
BEFORE UPDATE ON iam_permission_request_groups
FOR EACH ROW EXECUTE FUNCTION iam_guard_permission_request_terminal_v1();
ALTER TABLE iam_permission_request_groups
    ENABLE ALWAYS TRIGGER iam_permission_request_groups_terminal_guard;

CREATE TRIGGER iam_permission_requests_revision_guard
BEFORE UPDATE ON iam_permission_requests
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_permission_requests
    ENABLE ALWAYS TRIGGER iam_permission_requests_revision_guard;

CREATE TRIGGER iam_permission_requests_immutable_guard
BEFORE UPDATE ON iam_permission_requests
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'group_id', 'ordinal', 'target_kind', 'target_user_id',
    'target_role_id', 'target_organization_id', 'permission_id',
    'permission_match_kind', 'permission_is_protected',
    'planned_binding_id', 'created_at'
);
ALTER TABLE iam_permission_requests
    ENABLE ALWAYS TRIGGER iam_permission_requests_immutable_guard;

CREATE TRIGGER iam_permission_requests_state_guard
BEFORE UPDATE ON iam_permission_requests
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'PENDING>IN_REVIEW', 'PENDING>CANCELLED', 'PENDING>EXPIRED',
    'IN_REVIEW>APPROVED', 'IN_REVIEW>REJECTED',
    'IN_REVIEW>CANCELLED', 'IN_REVIEW>EXPIRED'
);
ALTER TABLE iam_permission_requests
    ENABLE ALWAYS TRIGGER iam_permission_requests_state_guard;

CREATE TRIGGER iam_permission_requests_terminal_guard
BEFORE UPDATE ON iam_permission_requests
FOR EACH ROW EXECUTE FUNCTION iam_guard_permission_request_terminal_v1();
ALTER TABLE iam_permission_requests
    ENABLE ALWAYS TRIGGER iam_permission_requests_terminal_guard;

CREATE CONSTRAINT TRIGGER iam_permission_request_groups_cardinality_guard
AFTER INSERT OR UPDATE ON iam_permission_request_groups
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION iam_guard_permission_request_group_v1();
ALTER TABLE iam_permission_request_groups
    ENABLE ALWAYS TRIGGER iam_permission_request_groups_cardinality_guard;

CREATE CONSTRAINT TRIGGER iam_permission_requests_group_guard
AFTER INSERT OR UPDATE OR DELETE ON iam_permission_requests
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION iam_guard_permission_request_group_v1();
ALTER TABLE iam_permission_requests
    ENABLE ALWAYS TRIGGER iam_permission_requests_group_guard;

CREATE TRIGGER iam_permission_request_groups_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_permission_request_groups
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_permission_request_groups
    ENABLE ALWAYS TRIGGER iam_permission_request_groups_removal_guard;

CREATE TRIGGER iam_permission_requests_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_permission_requests
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_permission_requests
    ENABLE ALWAYS TRIGGER iam_permission_requests_removal_guard;
