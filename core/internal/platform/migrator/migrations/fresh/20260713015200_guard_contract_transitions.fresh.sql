-- +goose Up
CREATE TRIGGER schema_contracts_validate_transition
BEFORE UPDATE OR DELETE ON platform.schema_contracts
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_validate_contract_transition();
