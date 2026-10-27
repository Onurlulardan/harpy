-- +goose Up
GRANT EXECUTE ON FUNCTION platform.cutover_record_billing_forms_rehearsal_v1(
    uuid,uuid,bigint,uuid,uuid,uuid,bigint,bigint,text,text,
    uuid,uuid,uuid,bigint,bigint,text,text,text,timestamptz
) TO harpy_adopter;

