-- +goose Up
CREATE TRIGGER tenancy_customer_provisioning_requests_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_customer_provisioning_requests
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();

ALTER TABLE platform.tenancy_customer_provisioning_requests
    ENABLE ALWAYS TRIGGER tenancy_customer_provisioning_requests_reject_mutation;
