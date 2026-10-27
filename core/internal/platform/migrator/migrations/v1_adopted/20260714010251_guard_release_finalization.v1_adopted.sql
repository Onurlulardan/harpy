-- +goose Up
CREATE CONSTRAINT TRIGGER release_generations_require_active_contract
AFTER INSERT ON platform.release_generations
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_validate_release_generation_finalization();
