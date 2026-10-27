-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_profile_snapshot_insert_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_schema_state text;
    v_field_key text;
    v_value_kind text;
    v_cardinality text;
    v_required boolean;
    v_classification text;
    v_normalization text;
    v_field_state text;
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM profile snapshot insert invocation is invalid';
    END IF;
    SELECT schema_version.state, field.field_key, field.value_kind,
           field.cardinality, field.required, field.classification,
           field.normalization_profile, field.state
    INTO v_schema_state, v_field_key, v_value_kind, v_cardinality,
         v_required, v_classification, v_normalization, v_field_state
    FROM iam_profile_schema_versions AS schema_version
    JOIN iam_profile_fields AS field
      ON field.id = NEW.field_id
     AND field.resource_kind = NEW.resource_kind
    WHERE schema_version.id = NEW.schema_version_id
      AND schema_version.resource_kind = NEW.resource_kind
    FOR KEY SHARE OF schema_version, field;
    IF NOT FOUND OR v_schema_state IS DISTINCT FROM 'draft' OR
       NEW.field_key IS DISTINCT FROM v_field_key OR
       NEW.value_kind IS DISTINCT FROM v_value_kind OR
       NEW.cardinality IS DISTINCT FROM v_cardinality OR
       NEW.required IS DISTINCT FROM v_required OR
       NEW.classification IS DISTINCT FROM v_classification OR
       NEW.normalization_profile IS DISTINCT FROM v_normalization OR
       NEW.field_state IS DISTINCT FROM v_field_state THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_field_revisions_snapshot_insert_guard',
            MESSAGE = 'IAM profile snapshot does not match its draft field';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_profile_snapshot_insert_v1() FROM PUBLIC;

CREATE TRIGGER iam_profile_field_revisions_insert_guard
BEFORE INSERT ON iam_profile_field_revisions
FOR EACH ROW EXECUTE FUNCTION iam_guard_profile_snapshot_insert_v1();
ALTER TABLE iam_profile_field_revisions
    ENABLE ALWAYS TRIGGER iam_profile_field_revisions_insert_guard;

CREATE TRIGGER iam_profile_field_revisions_update_guard
BEFORE UPDATE ON iam_profile_field_revisions
FOR EACH ROW EXECUTE FUNCTION iam_forbid_profile_destructive_mutation_v1();
ALTER TABLE iam_profile_field_revisions
    ENABLE ALWAYS TRIGGER iam_profile_field_revisions_update_guard;
