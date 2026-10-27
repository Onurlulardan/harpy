-- +goose Up
GRANT EXECUTE ON FUNCTION documents_v1_adoption_apply_and_record_v1(
    uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,
    pg_lsn,xid8,integer,text,jsonb,bytea,text,bigint,text
) TO harpy_adopter;
