-- +goose Up
REVOKE EXECUTE ON FUNCTION platform.billing_v1_base_adoption_assert_v1(
    uuid,uuid,uuid,text,text,uuid,bigint,text,timestamptz,bigint,text,text,text
) FROM forgecrud_adopter;
