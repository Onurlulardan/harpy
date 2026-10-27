-- +goose Up
GRANT SELECT ON iam_catalog_component_candidates TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_catalog_project_v2(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,text,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_catalog_component_apply_v2(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) TO harpy_runtime;
