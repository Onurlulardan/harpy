-- +goose Up
CREATE TRIGGER tenancy_schema_identity_reject_mutation
BEFORE UPDATE OR DELETE ON tenancy_schema_identity
FOR EACH ROW EXECUTE FUNCTION tenancy_reject_schema_identity_mutation();
