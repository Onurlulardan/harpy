-- +goose Up
CREATE TRIGGER extension_registry_reject_mutation
BEFORE UPDATE OR DELETE ON platform.extension_registry
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_reject_mutation();
