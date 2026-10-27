-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_version_evidence_immutable_v2()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='IAM catalog version evidence is immutable';
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_version_evidence_immutable_v2() FROM PUBLIC;

CREATE TRIGGER iam_catalog_version_evidence_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_evidence
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_immutable_v2();
CREATE TRIGGER iam_catalog_version_resources_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_resources
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_immutable_v2();
CREATE TRIGGER iam_catalog_version_grants_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_grants
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_immutable_v2();
CREATE TRIGGER iam_catalog_version_provisions_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_provisions
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_immutable_v2();
CREATE TRIGGER iam_catalog_version_operational_resources_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_operational_resources
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_immutable_v2();
CREATE TRIGGER iam_catalog_version_operational_grants_immutable
BEFORE UPDATE OR DELETE ON iam_catalog_version_operational_grants
FOR EACH ROW EXECUTE FUNCTION iam_catalog_version_evidence_immutable_v2();

ALTER TABLE iam_catalog_version_evidence
    ENABLE ALWAYS TRIGGER iam_catalog_version_evidence_immutable;
ALTER TABLE iam_catalog_version_resources
    ENABLE ALWAYS TRIGGER iam_catalog_version_resources_immutable;
ALTER TABLE iam_catalog_version_grants
    ENABLE ALWAYS TRIGGER iam_catalog_version_grants_immutable;
ALTER TABLE iam_catalog_version_provisions
    ENABLE ALWAYS TRIGGER iam_catalog_version_provisions_immutable;
ALTER TABLE iam_catalog_version_operational_resources
    ENABLE ALWAYS TRIGGER iam_catalog_version_operational_resources_immutable;
ALTER TABLE iam_catalog_version_operational_grants
    ENABLE ALWAYS TRIGGER iam_catalog_version_operational_grants_immutable;
