-- +goose Up
REVOKE ALL ON FUNCTION analytics_catalog_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_project_v2(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_catalog_component_apply_v2(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION analytics_catalog_project_v1(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_catalog_project_v2(
    name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,text,
    bytea,integer,integer,bytea,bytea,bytea,uuid,text,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_catalog_component_apply_v1(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_catalog_component_apply_v2(
    uuid,uuid,uuid,text,bigint,text,text,text,bigint,text,timestamptz
) TO harpy_runtime;

GRANT SELECT ON analytics_catalog_component_candidates
    TO harpy_runtime;
