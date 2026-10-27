-- +goose Up
CREATE TRIGGER release_generations_reject_mutation
BEFORE UPDATE OR DELETE ON platform.release_generations
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_reject_mutation();
