-- +goose Up
CREATE TRIGGER tenants_reject_insert_delete
INSTEAD OF INSERT OR DELETE ON public.tenants
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();
