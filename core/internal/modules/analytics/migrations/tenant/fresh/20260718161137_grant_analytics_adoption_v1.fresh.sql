-- +goose Up

GRANT EXECUTE ON FUNCTION analytics_adoption_record_item_v1(
    name,uuid,uuid,uuid,uuid,bytea,text,text,uuid,uuid,uuid,bytea,timestamptz
) TO forgecrud_adopter;
GRANT EXECUTE ON FUNCTION analytics_adoption_accept_v1(
    name,uuid,uuid,uuid,bigint,bigint,bigint,bigint,bytea,timestamptz
) TO forgecrud_adopter;
