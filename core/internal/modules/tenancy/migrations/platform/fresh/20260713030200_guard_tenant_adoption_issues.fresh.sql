-- +goose Up
CREATE TRIGGER tenant_adoption_issues_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenant_adoption_issues
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();
