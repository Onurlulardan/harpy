-- +goose Up
CREATE TRIGGER release_generations_validate_insert
BEFORE INSERT ON platform.release_generations
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_validate_release_generation_insert();
