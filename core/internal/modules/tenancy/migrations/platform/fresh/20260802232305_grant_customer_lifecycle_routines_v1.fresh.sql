-- +goose Up
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_list_v1(uuid,integer,uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_get_v1(uuid,uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_mutate_v1(
    uuid,uuid,uuid,uuid,text,text,bigint,text,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_accept_v1(
    uuid,uuid,uuid,uuid,text,bigint,text,bytea,uuid,uuid,uuid,bytea,
    timestamptz,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_read_v1(uuid,uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_claim_v1(uuid,timestamptz,integer)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_complete_offboard_v1(uuid,uuid,timestamptz)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_complete_purge_v1(uuid,uuid,uuid,bytea,timestamptz)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_lifecycle_fail_v1(uuid,uuid,text,timestamptz)
TO forgecrud_runtime;
