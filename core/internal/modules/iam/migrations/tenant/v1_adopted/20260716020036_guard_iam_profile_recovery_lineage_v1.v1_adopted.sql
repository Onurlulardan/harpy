-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_profile_recovery_lineage_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_parent_resource_kind text;
    v_parent_source_schema_id uuid;
    v_parent_target_schema_id uuid;
    v_parent_converter_version text;
    v_parent_plan_sha256 bytea;
    v_parent_root_plan_id uuid;
    v_parent_sequence bigint;
    v_parent_state text;
    v_target_definition_sha256 bytea;
    v_expected_root_plan_id uuid;
BEGIN
    IF TG_OP IN ('DELETE', 'TRUNCATE') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM profile backfill plan evidence is immutable';
    END IF;
    IF TG_OP = 'UPDATE' THEN
        IF OLD.state IN ('activated', 'fenced') AND NEW IS DISTINCT FROM OLD THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'Terminal IAM profile backfill evidence is immutable';
        END IF;
        IF NEW.id IS DISTINCT FROM OLD.id OR
           NEW.resource_kind IS DISTINCT FROM OLD.resource_kind OR
           NEW.source_schema_id IS DISTINCT FROM OLD.source_schema_id OR
           NEW.target_schema_id IS DISTINCT FROM OLD.target_schema_id OR
           NEW.converter_version IS DISTINCT FROM OLD.converter_version OR
           NEW.plan_sha256 IS DISTINCT FROM OLD.plan_sha256 OR
           NEW.source_count IS DISTINCT FROM OLD.source_count OR
           NEW.expected_source_sha256 IS DISTINCT FROM OLD.expected_source_sha256 OR
           NEW.expected_target_sha256 IS DISTINCT FROM OLD.expected_target_sha256 OR
           NEW.page_size IS DISTINCT FROM OLD.page_size OR
           NEW.page_count IS DISTINCT FROM OLD.page_count OR
           NEW.created_at IS DISTINCT FROM OLD.created_at OR
           NEW.recovery_of_plan_id IS DISTINCT FROM OLD.recovery_of_plan_id OR
           NEW.recovery_root_plan_id IS DISTINCT FROM OLD.recovery_root_plan_id OR
           NEW.recovery_sequence IS DISTINCT FROM OLD.recovery_sequence OR
           NEW.parent_plan_sha256 IS DISTINCT FROM OLD.parent_plan_sha256 OR
           NEW.target_definition_sha256 IS DISTINCT FROM OLD.target_definition_sha256 OR
           NEW.inherited_target_count IS DISTINCT FROM OLD.inherited_target_count OR
           NEW.inherited_target_sha256 IS DISTINCT FROM OLD.inherited_target_sha256 OR
           NEW.remaining_source_count IS DISTINCT FROM OLD.remaining_source_count OR
           NEW.remaining_source_sha256 IS DISTINCT FROM OLD.remaining_source_sha256 OR
           NEW.expected_remaining_target_sha256 IS DISTINCT FROM
               OLD.expected_remaining_target_sha256 OR
           NEW.initial_rolling_sha256 IS DISTINCT FROM OLD.initial_rolling_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'IAM profile backfill plan identity is immutable';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.recovery_of_plan_id IS NULL THEN
        RETURN NEW;
    END IF;
    SELECT parent.resource_kind, parent.source_schema_id, parent.target_schema_id,
           parent.converter_version, parent.plan_sha256,
           parent.recovery_root_plan_id, parent.recovery_sequence, parent.state,
           target.definition_sha256
    INTO v_parent_resource_kind, v_parent_source_schema_id,
         v_parent_target_schema_id, v_parent_converter_version,
         v_parent_plan_sha256, v_parent_root_plan_id, v_parent_sequence,
         v_parent_state, v_target_definition_sha256
    FROM iam_profile_backfill_plans AS parent
    JOIN iam_profile_schema_versions AS source
      ON source.id = parent.source_schema_id
     AND source.resource_kind = parent.resource_kind
     AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = parent.target_schema_id
     AND target.resource_kind = parent.resource_kind
     AND target.state = 'draft'
    WHERE parent.id = NEW.recovery_of_plan_id
    FOR SHARE OF parent, source, target;
    v_expected_root_plan_id := COALESCE(
        v_parent_root_plan_id, NEW.recovery_of_plan_id
    );
    IF v_parent_state IS DISTINCT FROM 'fenced' OR
       NEW.state IS DISTINCT FROM 'planned' OR
       NEW.resource_kind IS DISTINCT FROM v_parent_resource_kind OR
       NEW.source_schema_id IS DISTINCT FROM v_parent_source_schema_id OR
       NEW.target_schema_id IS DISTINCT FROM v_parent_target_schema_id OR
       NEW.converter_version IS DISTINCT FROM v_parent_converter_version OR
       NEW.parent_plan_sha256 IS DISTINCT FROM v_parent_plan_sha256 OR
       NEW.target_definition_sha256 IS DISTINCT FROM v_target_definition_sha256 OR
       NEW.recovery_root_plan_id IS DISTINCT FROM v_expected_root_plan_id OR
       NEW.recovery_sequence IS DISTINCT FROM COALESCE(v_parent_sequence, 0) + 1 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_backfill_plans_recovery_lineage_guard',
            MESSAGE = 'IAM profile recovery lineage is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION iam_guard_profile_recovery_lineage_v1() FROM PUBLIC;

CREATE TRIGGER iam_profile_backfill_plans_recovery_lineage_guard
BEFORE INSERT OR UPDATE OR DELETE ON iam_profile_backfill_plans
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_recovery_lineage_v1();

CREATE TRIGGER iam_profile_backfill_plans_recovery_truncate_guard
BEFORE TRUNCATE ON iam_profile_backfill_plans
FOR EACH STATEMENT EXECUTE FUNCTION iam_guard_profile_recovery_lineage_v1();

ALTER TABLE iam_profile_backfill_plans
    ENABLE ALWAYS TRIGGER iam_profile_backfill_plans_recovery_lineage_guard;
ALTER TABLE iam_profile_backfill_plans
    ENABLE ALWAYS TRIGGER iam_profile_backfill_plans_recovery_truncate_guard;

CREATE FUNCTION iam_assert_profile_recovery_materialization_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NEW.recovery_of_plan_id IS NULL THEN
        RETURN NULL;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM (
            SELECT * FROM iam_profile_backfill_mappings
            WHERE plan_id = NEW.recovery_of_plan_id
        ) AS parent
        FULL JOIN (
            SELECT * FROM iam_profile_backfill_mappings
            WHERE plan_id = NEW.id
        ) AS child ON child.ordinal = parent.ordinal
        WHERE parent.ordinal IS NULL OR child.ordinal IS NULL OR
              child.source_field_id <> parent.source_field_id OR
              child.target_field_id <> parent.target_field_id OR
              child.converter_kind <> parent.converter_kind OR
              child.semantic_sha256 <> parent.semantic_sha256
    ) OR NEW.page_count <> (
        SELECT pg_catalog.count(*) FROM iam_profile_backfill_pages AS page
        WHERE page.plan_id = NEW.id
    ) OR NEW.source_count <> COALESCE((
        SELECT pg_catalog.sum(page.expected_count)
        FROM iam_profile_backfill_pages AS page WHERE page.plan_id = NEW.id
    ), 0) OR NOT EXISTS (
        SELECT 1 FROM iam_profile_backfill_checkpoints AS checkpoint
        WHERE checkpoint.plan_id = NEW.id AND checkpoint.completed_pages = 0
          AND checkpoint.converted_count = 0 AND checkpoint.error_count = 0
          AND checkpoint.rolling_sha256 = NEW.initial_rolling_sha256
          AND checkpoint.last_page_number IS NULL AND checkpoint.revision = 1
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_recovery_materialization_guard',
            MESSAGE = 'IAM profile recovery materialization is incomplete';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_assert_profile_recovery_materialization_v1()
    FROM PUBLIC;

CREATE CONSTRAINT TRIGGER iam_profile_recovery_materialization_guard
AFTER INSERT ON iam_profile_backfill_plans
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION iam_assert_profile_recovery_materialization_v1();

ALTER TABLE iam_profile_backfill_plans ENABLE ALWAYS TRIGGER
    iam_profile_recovery_materialization_guard;
