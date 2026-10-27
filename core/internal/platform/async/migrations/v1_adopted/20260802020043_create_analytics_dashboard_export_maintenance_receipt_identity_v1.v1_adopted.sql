-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_export_sha1_v1(p_value bytea)
RETURNS bytea
LANGUAGE plpgsql IMMUTABLE STRICT SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_h0 bigint:=1732584193;
    v_h1 bigint:=4023233417;
    v_h2 bigint:=2562383102;
    v_h3 bigint:=271733878;
    v_h4 bigint:=3285377520;
    v_a bigint;
    v_b bigint;
    v_c bigint;
    v_d bigint;
    v_e bigint;
    v_f bigint;
    v_k bigint;
    v_temp bigint;
    v_word bigint;
    v_words bigint[]:=pg_catalog.array_fill(0::bigint,ARRAY[80]);
    v_data bytea;
    v_length integer:=pg_catalog.octet_length(p_value);
    v_padded integer;
    v_block integer;
    v_round integer;
    v_mask constant bigint:=4294967295;
BEGIN
    IF v_length>4096 THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='SHA-1 input is too large';
    END IF;
    v_padded:=((v_length+72)/64)*64;
    v_data:=p_value||pg_catalog.decode('80','hex')||pg_catalog.decode(
        pg_catalog.repeat('00',v_padded-v_length-9),'hex')||
        pg_catalog.int8send(v_length::bigint*8);
    FOR v_block IN 0..(v_padded/64)-1 LOOP
        FOR v_round IN 0..15 LOOP
            v_words[v_round+1]:=
                (pg_catalog.get_byte(v_data,v_block*64+v_round*4)::bigint<<24)|
                (pg_catalog.get_byte(v_data,v_block*64+v_round*4+1)::bigint<<16)|
                (pg_catalog.get_byte(v_data,v_block*64+v_round*4+2)::bigint<<8)|
                 pg_catalog.get_byte(v_data,v_block*64+v_round*4+3)::bigint;
        END LOOP;
        FOR v_round IN 16..79 LOOP
            v_word:=v_words[v_round-2]#v_words[v_round-7]#
                v_words[v_round-13]#v_words[v_round-15];
            v_words[v_round+1]:=((v_word<<1)&v_mask)|(v_word>>31);
        END LOOP;
        v_a:=v_h0; v_b:=v_h1; v_c:=v_h2; v_d:=v_h3; v_e:=v_h4;
        FOR v_round IN 0..79 LOOP
            IF v_round<20 THEN
                v_f:=(v_b&v_c)|((v_b#v_mask)&v_d); v_k:=1518500249;
            ELSIF v_round<40 THEN
                v_f:=v_b#v_c#v_d; v_k:=1859775393;
            ELSIF v_round<60 THEN
                v_f:=(v_b&v_c)|(v_b&v_d)|(v_c&v_d); v_k:=2400959708;
            ELSE
                v_f:=v_b#v_c#v_d; v_k:=3395469782;
            END IF;
            v_temp:=((v_a<<5)&v_mask)|(v_a>>27);
            v_temp:=(v_temp+v_f+v_e+v_k+v_words[v_round+1])&v_mask;
            v_e:=v_d; v_d:=v_c;
            v_c:=(((v_b<<30)&v_mask)|(v_b>>2))&v_mask;
            v_b:=v_a; v_a:=v_temp;
        END LOOP;
        v_h0:=(v_h0+v_a)&v_mask; v_h1:=(v_h1+v_b)&v_mask;
        v_h2:=(v_h2+v_c)&v_mask; v_h3:=(v_h3+v_d)&v_mask;
        v_h4:=(v_h4+v_e)&v_mask;
    END LOOP;
    RETURN pg_catalog.decode(
        pg_catalog.lpad(pg_catalog.to_hex(v_h0),8,'0')||
        pg_catalog.lpad(pg_catalog.to_hex(v_h1),8,'0')||
        pg_catalog.lpad(pg_catalog.to_hex(v_h2),8,'0')||
        pg_catalog.lpad(pg_catalog.to_hex(v_h3),8,'0')||
        pg_catalog.lpad(pg_catalog.to_hex(v_h4),8,'0'),'hex');
END;
$function$;

CREATE FUNCTION platform.platform_async_analytics_export_maint_receipt_id_v1(
    p_effect_id uuid,p_kind text
)
RETURNS uuid
LANGUAGE plpgsql IMMUTABLE STRICT SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_code text;
    v_digest bytea;
    v_uuid bytea;
    v_hex text;
BEGIN
    v_code:=CASE p_kind WHEN 'analytics.dashboard.export.delete' THEN
        'ANALYTICS_EXPORT_DELETED' WHEN 'analytics.dashboard.export.expire' THEN
        'ANALYTICS_EXPORT_EXPIRY_COMPLETE' END;
    IF v_code IS NULL OR pg_catalog.uuid_extract_version(p_effect_id)<>5 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export receipt identity input is invalid';
    END IF;
    v_digest:=platform.platform_async_analytics_export_sha1_v1(
        pg_catalog.uuid_send('8e7a6d02-df6d-4327-b179-160296b50b8e'::uuid)||
        pg_catalog.convert_to(p_effect_id::text,'UTF8')||
        pg_catalog.decode('00','hex')||pg_catalog.convert_to(v_code,'UTF8'));
    v_uuid:=pg_catalog.substr(v_digest,1,16);
    v_uuid:=pg_catalog.set_byte(v_uuid,6,
        (pg_catalog.get_byte(v_uuid,6)&15)|80);
    v_uuid:=pg_catalog.set_byte(v_uuid,8,
        (pg_catalog.get_byte(v_uuid,8)&63)|128);
    v_hex:=pg_catalog.encode(v_uuid,'hex');
    RETURN (pg_catalog.substr(v_hex,1,8)||'-'||pg_catalog.substr(v_hex,9,4)||'-'||
        pg_catalog.substr(v_hex,13,4)||'-'||pg_catalog.substr(v_hex,17,4)||'-'||
        pg_catalog.substr(v_hex,21,12))::uuid;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_analytics_export_sha1_v1(bytea)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_export_maint_receipt_id_v1(uuid,text)
FROM PUBLIC;
