-- +goose Up
CREATE TRIGGER bootstrap_receipts_reject_mutation
BEFORE UPDATE OR DELETE ON platform.bootstrap_receipts
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_reject_mutation();
