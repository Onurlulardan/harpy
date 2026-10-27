-- +goose Up
REVOKE EXECUTE ON FUNCTION system_resource_layout_field_reference_evidence_v1(
    name,text,uuid
) FROM harpy_runtime;

GRANT EXECUTE ON FUNCTION system_resource_layout_field_reference_evidence_v2(
    name,text,uuid
) TO harpy_runtime;
