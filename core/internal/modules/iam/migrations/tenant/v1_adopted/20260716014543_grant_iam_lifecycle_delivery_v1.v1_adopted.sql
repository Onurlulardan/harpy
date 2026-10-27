-- +goose Up
GRANT EXECUTE ON FUNCTION iam_email_verification_delivery_target_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_password_reset_delivery_target_v1(
    name,text,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_setup_delivery_target_v1(
    name,uuid,uuid,uuid,bigint,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_email_verification_start_v2(
    name,uuid,uuid,bytea,uuid,uuid,bytea,text,bytea,bytea,
    uuid,uuid,bigint,bigint,uuid,uuid,uuid,uuid,text,text,
    bytea,bytea,bytea,bytea,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_password_reset_start_v2(
    name,uuid,uuid,bytea,uuid,text,bytea,bytea,text,bytea,bytea,
    uuid,uuid,uuid,uuid,text,text,bytea,bytea,bytea,bytea,bytea
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION iam_portal_setup_start_v2(
    name,uuid,uuid,bytea,uuid,uuid,bytea,text,bytea,bytea,
    uuid,uuid,bigint,bigint,uuid,uuid,uuid,uuid,text,text,
    bytea,bytea,bytea,bytea,bytea
) TO harpy_runtime;
