-- +goose Up
REVOKE ALL ON FUNCTION platform.platformauthority_reject_mutation_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_guard_operation_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_guard_head_v1() FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_guard_binding_v2() FROM PUBLIC;

REVOKE ALL ON FUNCTION platform.platformauthority_bind_v2(
    uuid,uuid,text,text,uuid,text,text,bigint,text,text,text,bytea,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_get_binding_v2(uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_prepare_v2(
    uuid,uuid,text,text,uuid,bigint,text,text,bytea,text,bytea,text,
    timestamptz,timestamptz,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_rotate_signer_v1(
    uuid,uuid,bigint,text,bigint,text,text,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_fence_signer_v1(
    uuid,bigint,text,text,timestamptz
) FROM PUBLIC;

REVOKE EXECUTE ON FUNCTION platform.platformauthority_bind_v1(
    uuid,uuid,text,text,uuid,text,text,bigint,text,text,text,bytea,timestamptz
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION platform.platformauthority_get_binding_v1(uuid)
FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION platform.platformauthority_prepare_v1(
    uuid,uuid,text,text,uuid,bigint,text,text,bytea,text,bytea,text,
    timestamptz,timestamptz,timestamptz
) FROM forgecrud_runtime;

GRANT EXECUTE ON FUNCTION platform.platformauthority_bind_v2(
    uuid,uuid,text,text,uuid,text,text,bigint,text,text,text,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_get_binding_v2(uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_prepare_v2(
    uuid,uuid,text,text,uuid,bigint,text,text,bytea,text,bytea,text,
    timestamptz,timestamptz,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_rotate_signer_v1(
    uuid,uuid,bigint,text,bigint,text,text,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platformauthority_fence_signer_v1(
    uuid,bigint,text,text,timestamptz
) TO forgecrud_runtime;
