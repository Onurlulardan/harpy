-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_evidence_immutable_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='55000',
        MESSAGE='Analytics V1 adoption evidence is immutable';
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER analytics_v1_adoption_differentials_immutable
BEFORE UPDATE OR DELETE ON analytics_v1_adoption_differentials
FOR EACH STATEMENT EXECUTE FUNCTION analytics_v1_adoption_evidence_immutable_v1();
CREATE TRIGGER analytics_v1_adoption_receipts_immutable
BEFORE UPDATE OR DELETE ON analytics_v1_adoption_receipts
FOR EACH STATEMENT EXECUTE FUNCTION analytics_v1_adoption_evidence_immutable_v1();

ALTER TABLE analytics_v1_adoption_differentials
    ENABLE ALWAYS TRIGGER analytics_v1_adoption_differentials_immutable;
ALTER TABLE analytics_v1_adoption_receipts
    ENABLE ALWAYS TRIGGER analytics_v1_adoption_receipts_immutable;

REVOKE ALL ON FUNCTION analytics_v1_adoption_evidence_immutable_v1() FROM PUBLIC;
