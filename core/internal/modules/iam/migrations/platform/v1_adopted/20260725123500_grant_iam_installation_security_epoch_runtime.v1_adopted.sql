-- +goose Up
REVOKE ALL ON FUNCTION platform.iam_security_epoch_initialize_v1(
    uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_security_epoch_initialize_v1(
    uuid,timestamptz
) TO harpy_runtime;
