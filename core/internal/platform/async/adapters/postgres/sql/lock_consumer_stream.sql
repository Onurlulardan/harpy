SELECT
    platform.platform_async_assert_current_authority_v1(
        $1::text, $2::uuid, $3::text, $4::bigint, $5::text, $6::bigint
    ),
    pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        pg_catalog.concat_ws(E'\x1f',
            'platform_async_inbox_v1', $7::text, $8::integer::text,
            $1::text, $2::uuid::text, $9::text, $10::uuid::text
        ), 0
    ));
