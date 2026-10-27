-- +goose Up
CREATE TRIGGER cutover_batch_receipts_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_batch_receipts
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();

CREATE TRIGGER cutover_contract_probe_rows_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_contract_probe_rows
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
