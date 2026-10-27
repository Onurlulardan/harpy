-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reject_profile_preconversion_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'IAM profile preconversion evidence is immutable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reject_profile_preconversion_mutation_v1()
    FROM PUBLIC;

CREATE TRIGGER iam_profile_preconversions_mutation_guard
BEFORE UPDATE OR DELETE ON iam_profile_backfill_preconversions
FOR EACH ROW EXECUTE FUNCTION iam_reject_profile_preconversion_mutation_v1();
ALTER TABLE iam_profile_backfill_preconversions ENABLE ALWAYS TRIGGER
    iam_profile_preconversions_mutation_guard;

CREATE TRIGGER iam_profile_preconversions_truncate_guard
BEFORE TRUNCATE ON iam_profile_backfill_preconversions
FOR EACH STATEMENT EXECUTE FUNCTION iam_reject_profile_preconversion_mutation_v1();
ALTER TABLE iam_profile_backfill_preconversions ENABLE ALWAYS TRIGGER
    iam_profile_preconversions_truncate_guard;
