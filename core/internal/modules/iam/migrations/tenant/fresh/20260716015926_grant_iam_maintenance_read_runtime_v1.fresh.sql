-- +goose Up
GRANT EXECUTE ON FUNCTION iam_maintenance_directory_repair_page_v1(
    name,uuid,uuid,uuid,integer,text,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_directory_checksum_sources_v1(
    name,uuid,uuid,uuid,integer,text,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_page_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_mappings_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_schema_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,uuid
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_maintenance_profile_resources_read_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,integer
) TO harpy_runtime;
