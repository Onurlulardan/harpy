-- +goose Up
REVOKE ALL ON FUNCTION platform.platformauthority_bind_v1(
    uuid,uuid,text,text,uuid,text,text,bigint,text,text,text,bytea,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_get_binding_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_prepare_v1(
    uuid,uuid,text,text,uuid,bigint,text,text,bytea,text,bytea,text,timestamptz,timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_record_ack_v1(
    uuid,uuid,text,uuid,bigint,text,text,text,text,bytea,bigint,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_commit_v1(uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_get_operation_v1(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_get_head_v1(uuid,text,uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_bind_v1(
    uuid,uuid,text,text,uuid,text,text,bigint,text,text,text,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_get_binding_v1(uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_prepare_v1(
    uuid,uuid,text,text,uuid,bigint,text,text,bytea,text,bytea,text,timestamptz,timestamptz,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_record_ack_v1(
    uuid,uuid,text,uuid,bigint,text,text,text,text,bytea,bigint,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_commit_v1(uuid,timestamptz) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_get_operation_v1(uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_get_head_v1(uuid,text,uuid) TO harpy_runtime;
