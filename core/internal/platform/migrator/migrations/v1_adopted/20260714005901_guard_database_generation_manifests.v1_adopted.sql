-- +goose Up
CREATE TRIGGER database_generation_manifests_reject_mutation
BEFORE UPDATE OR DELETE ON platform.database_generation_manifests
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_reject_mutation();
