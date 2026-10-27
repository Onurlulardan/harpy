-- +goose Up
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_accept_v1(
    uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,
    uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,
    text,text,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_claim_v1(
    uuid,uuid,text,text,bigint,integer
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_record_v1(
    uuid,uuid,uuid,text,text,text,timestamptz,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_authority_ack_v1(
    uuid,uuid,text,uuid,bigint,text,text,text,text,bigint,timestamptz,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v1(
    uuid,text,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_complete_v1(
    uuid,uuid,text,text,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_reconcile_v1(
    uuid,uuid,text,text,bigint
) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_accept_v1(
    uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,
    uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,
    text,text,text,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_claim_v1(
    uuid,uuid,text,text,bigint,integer
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_record_v1(
    uuid,uuid,uuid,text,text,text,timestamptz,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_authority_ack_v1(
    uuid,uuid,text,uuid,bigint,text,text,text,text,bigint,timestamptz,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v1(
    uuid,text,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_complete_v1(
    uuid,uuid,text,text,bigint
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_reconcile_v1(
    uuid,uuid,text,text,bigint
) TO harpy_runtime;
