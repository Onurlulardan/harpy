-- +goose Up
CREATE TRIGGER tenancy_policy_initialization_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_deployment_policy_initializations
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();
