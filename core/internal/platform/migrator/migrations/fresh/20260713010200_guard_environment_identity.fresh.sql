-- +goose Up
CREATE TRIGGER environment_identity_reject_mutation
BEFORE UPDATE OR DELETE ON platform.environment_identity
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_reject_mutation();
