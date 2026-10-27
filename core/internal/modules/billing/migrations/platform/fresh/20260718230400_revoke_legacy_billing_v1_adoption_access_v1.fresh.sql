-- +goose Up
REVOKE EXECUTE ON FUNCTION platform.billing_convert_v1_storage_megabytes_v1(bigint)
    FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION platform.billing_register_v1_adoption_v1(
    uuid, uuid, uuid, bytea, text, bytea, text
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION platform.billing_apply_v1_adoption_v1(uuid, uuid)
    FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION platform.billing_verify_v1_adoption_v1(uuid, uuid)
    FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION platform.billing_rehearse_v1_adoption_rollback_v1(
    uuid, uuid, uuid, text, timestamptz
) FROM harpy_adopter;
REVOKE EXECUTE ON FUNCTION platform.billing_assert_v1_adoption_ready_v1(
    uuid, uuid, text
) FROM harpy_adopter;
