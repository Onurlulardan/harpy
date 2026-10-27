-- +goose Up
GRANT EXECUTE ON FUNCTION iam_profile_historical_read_v1(name,uuid)
    TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_active_catalog_read_v1(name,text)
    TO harpy_runtime;

GRANT EXECUTE ON FUNCTION iam_profile_document_read_v1(name,text,uuid)
    TO harpy_runtime;
