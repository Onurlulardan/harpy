-- +goose Up
GRANT EXECUTE ON FUNCTION iam_owner_transfer_prepare_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,bigint,bytea,timestamptz,text,
    boolean,boolean,boolean,uuid,bigint,bigint,bigint,uuid,uuid,uuid,bigint,
    bigint,bytea,bytea,bytea,text,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_owner_transfer_ack_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,bytea,timestamptz,text,
    boolean,boolean,boolean,uuid,uuid,bigint,bigint,bytea,bytea,bytea,
    bytea,timestamptz,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_owner_transfer_commit_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,uuid,bytea
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_owner_transfer_compensate_prepare_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,bytea,timestamptz,text,
    boolean,boolean,boolean,uuid,bytea,bigint,bytea,text,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_owner_transfer_compensate_apply_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,bytea,timestamptz,text,
    boolean,boolean,boolean,uuid,uuid,bigint,bigint,bytea,bytea,bytea,
    bytea,timestamptz,text
) TO forgecrud_runtime;
