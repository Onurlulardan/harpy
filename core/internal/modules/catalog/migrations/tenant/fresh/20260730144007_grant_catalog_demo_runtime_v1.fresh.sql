-- +goose Up
GRANT EXECUTE ON FUNCTION catalog_demo_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,
    bytea,integer,integer,bytea,text,bytea,timestamptz
) TO harpy_runtime;
