-- +goose Up
GRANT EXECUTE ON FUNCTION iam_platform_operator_grant_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,bytea,timestamptz,text,
    boolean,boolean,boolean,uuid,bigint,uuid,uuid,uuid,uuid,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_platform_operator_revoke_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,bytea,timestamptz,text,
    boolean,boolean,boolean,uuid,bigint,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_platform_authorize_v1(
    name,uuid,bigint,bigint,bigint,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION iam_platform_session_context_v1(
    name,uuid,bigint,bigint
) TO forgecrud_runtime;
