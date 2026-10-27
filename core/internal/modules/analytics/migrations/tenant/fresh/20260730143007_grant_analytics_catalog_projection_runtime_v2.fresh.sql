-- +goose Up
GRANT EXECUTE ON FUNCTION analytics_catalog_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) TO forgecrud_runtime;

GRANT SELECT ON analytics_catalog_definitions TO forgecrud_runtime;
