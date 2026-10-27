-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_profile_recovery_activation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_total_count bigint;
    v_target_count bigint;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       NEW.recovery_of_plan_id IS NULL OR NEW.state IS DISTINCT FROM 'activated' OR
       OLD.state IS NOT DISTINCT FROM 'activated' THEN
        RETURN NEW;
    END IF;
    IF OLD.state NOT IN ('planned','running') OR
       NOT EXISTS (
           SELECT 1
           FROM iam_profile_backfill_checkpoints AS checkpoint
           WHERE checkpoint.plan_id = NEW.id
             AND checkpoint.completed_pages = NEW.page_count
             AND checkpoint.converted_count = NEW.remaining_source_count
             AND checkpoint.error_count = 0
             AND checkpoint.rolling_sha256 =
                 NEW.expected_remaining_target_sha256
             AND ((NEW.page_count = 0 AND checkpoint.last_page_number IS NULL) OR
                  checkpoint.last_page_number = NEW.page_count)
       ) OR EXISTS (
           SELECT 1 FROM iam_profile_backfill_pages AS page
           WHERE page.plan_id = NEW.id AND (
               page.state <> 'completed' OR page.error_count <> 0 OR
               page.converted_count <> page.expected_count OR
               page.result_sha256 <> page.expected_target_sha256
           )
       ) OR NOT EXISTS (
           SELECT 1
           FROM iam_authorization_operation_receipts AS receipt
           JOIN iam_authorization_clock AS clock ON clock.singleton
           WHERE receipt.command_key = 'iam.profile_backfill.recovery_create'
             AND receipt.result_id = NEW.id AND receipt.result_revision = 1
             AND receipt.result_sha256 = NEW.plan_sha256
             AND receipt.clock_value <= clock.clock_value
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_recovery_activation_evidence_guard',
            MESSAGE = 'IAM profile recovery activation evidence is incomplete';
    END IF;
    SELECT pg_catalog.count(*), pg_catalog.count(*) FILTER (
               WHERE resource.schema_version_id = NEW.target_schema_id
           )
    INTO v_total_count, v_target_count
    FROM iam_profile_recovery_resources_v1 AS resource
    WHERE resource.resource_kind = NEW.resource_kind;
    IF v_total_count < NEW.inherited_target_count + NEW.remaining_source_count OR
       v_target_count <> v_total_count OR NOT EXISTS (
           SELECT 1 FROM iam_profile_schema_versions AS source
           JOIN iam_profile_schema_versions AS target
             ON target.id = NEW.target_schema_id
            AND target.resource_kind = NEW.resource_kind
            AND target.state = 'active'
           WHERE source.id = NEW.source_schema_id
             AND source.resource_kind = NEW.resource_kind
             AND source.state = 'retired'
             AND target.version = source.version + 1
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_recovery_activation_lineage_guard',
            MESSAGE = 'IAM profile recovery activation lineage is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_profile_recovery_activation_v1()
    FROM PUBLIC;

CREATE TRIGGER iam_profile_backfill_plans_recovery_activation_guard
BEFORE UPDATE ON iam_profile_backfill_plans
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_recovery_activation_v1();

ALTER TABLE iam_profile_backfill_plans ENABLE ALWAYS TRIGGER
    iam_profile_backfill_plans_recovery_activation_guard;
