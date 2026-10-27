-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_forbid_profile_destructive_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        CONSTRAINT = TG_TABLE_NAME || '_destructive_mutation_guard',
        MESSAGE = 'IAM profile snapshot removal or rewrite is forbidden';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_forbid_profile_destructive_mutation_v1()
    FROM PUBLIC;

CREATE TRIGGER iam_profile_schema_versions_delete_guard
BEFORE DELETE ON iam_profile_schema_versions
FOR EACH ROW EXECUTE FUNCTION iam_forbid_profile_destructive_mutation_v1();
ALTER TABLE iam_profile_schema_versions
    ENABLE ALWAYS TRIGGER iam_profile_schema_versions_delete_guard;

CREATE TRIGGER iam_profile_schema_versions_truncate_guard
BEFORE TRUNCATE ON iam_profile_schema_versions
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_profile_destructive_mutation_v1();
ALTER TABLE iam_profile_schema_versions
    ENABLE ALWAYS TRIGGER iam_profile_schema_versions_truncate_guard;

CREATE TRIGGER iam_profile_fields_delete_guard
BEFORE DELETE ON iam_profile_fields
FOR EACH ROW EXECUTE FUNCTION iam_forbid_profile_destructive_mutation_v1();
ALTER TABLE iam_profile_fields
    ENABLE ALWAYS TRIGGER iam_profile_fields_delete_guard;

CREATE TRIGGER iam_profile_fields_truncate_guard
BEFORE TRUNCATE ON iam_profile_fields
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_profile_destructive_mutation_v1();
ALTER TABLE iam_profile_fields
    ENABLE ALWAYS TRIGGER iam_profile_fields_truncate_guard;

CREATE TRIGGER iam_profile_field_revisions_delete_guard
BEFORE DELETE ON iam_profile_field_revisions
FOR EACH ROW EXECUTE FUNCTION iam_forbid_profile_destructive_mutation_v1();
ALTER TABLE iam_profile_field_revisions
    ENABLE ALWAYS TRIGGER iam_profile_field_revisions_delete_guard;

CREATE TRIGGER iam_profile_field_revisions_truncate_guard
BEFORE TRUNCATE ON iam_profile_field_revisions
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_profile_destructive_mutation_v1();
ALTER TABLE iam_profile_field_revisions
    ENABLE ALWAYS TRIGGER iam_profile_field_revisions_truncate_guard;
