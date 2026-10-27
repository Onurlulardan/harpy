-- +goose Up
GRANT EXECUTE ON FUNCTION iam_profile_historical_read_v1(name,uuid)
    TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_active_catalog_read_v1(name,text)
    TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_document_read_v1(name,text,uuid)
    TO forgecrud_runtime;
