-- +goose Up
REVOKE EXECUTE ON FUNCTION analytics_adoption_record_item_v1(
    name,uuid,uuid,uuid,uuid,bytea,text,text,uuid,uuid,uuid,bytea,timestamptz
) FROM forgecrud_adopter;
REVOKE EXECUTE ON FUNCTION analytics_adoption_accept_v1(
    name,uuid,uuid,uuid,bigint,bigint,bigint,bigint,bytea,timestamptz
) FROM forgecrud_adopter;

GRANT EXECUTE ON FUNCTION analytics_v1_adoption_apply_and_record_v1(
    uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,
    pg_lsn,xid8,integer,text,jsonb,bytea,text,bigint,text
) TO forgecrud_adopter;
