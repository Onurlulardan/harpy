#!/usr/bin/env bash
set -euo pipefail

required=(
  POSTGRES_PASSWORD PG_READY_HOST PG_READY_USER PG_READY_DATABASE
  PG_READY_SERVER_VERSION_NUM PG_READY_DATA_CHECKSUMS PG_READY_DATABASE_OWNER
  PG_READY_ENCODING PG_READY_LOCALE_PROVIDER PG_READY_COLLATION PG_READY_CTYPE
  PG_READY_VECTOR_VERSION PG_READY_VECTOR_SCHEMA PG_READY_REQUIRE_SOURCE_IDENTITY
)
for name in "${required[@]}"; do
  [[ -n "${!name:-}" ]] || { printf '%s is required\n' "$name" >&2; exit 64; }
done

export PGPASSWORD="$POSTGRES_PASSWORD"
export PGOPTIONS='-c default_transaction_read_only=on -c statement_timeout=10000'
row=$(psql \
  --host="$PG_READY_HOST" \
  --port=5432 \
  --username="$PG_READY_USER" \
  --dbname="$PG_READY_DATABASE" \
  --no-psqlrc --quiet --tuples-only --no-align --field-separator='|' \
  --command="
    SELECT current_database(),
           current_setting('server_version_num'),
           current_setting('data_checksums'),
           pg_encoding_to_char(database.encoding),
           database.datlocprovider::text,
           database.datcollate,
           database.datctype,
           pg_get_userbyid(database.datdba),
           control.system_identifier::text,
           database.oid,
           extension.extversion,
           namespace.nspname
    FROM pg_database AS database
    CROSS JOIN pg_control_system() AS control
    JOIN pg_extension AS extension ON extension.extname = 'vector'
    JOIN pg_namespace AS namespace ON namespace.oid = extension.extnamespace
    WHERE database.datname = current_database();")

IFS='|' read -r database_name version checksums encoding provider collation ctype owner \
  system_identifier database_oid vector_version vector_schema <<<"$row"
[[ "$database_name" == "$PG_READY_DATABASE" ]]
[[ "$version" == "$PG_READY_SERVER_VERSION_NUM" ]]
[[ "$checksums" == "$PG_READY_DATA_CHECKSUMS" ]]
[[ "$encoding" == "$PG_READY_ENCODING" ]]
[[ "$provider" == "$PG_READY_LOCALE_PROVIDER" ]]
[[ "$collation" == "$PG_READY_COLLATION" ]]
[[ "$ctype" == "$PG_READY_CTYPE" ]]
[[ "$owner" == "$PG_READY_DATABASE_OWNER" ]]
[[ "$vector_version" == "$PG_READY_VECTOR_VERSION" ]]
[[ "$vector_schema" == "$PG_READY_VECTOR_SCHEMA" ]]
[[ -n "$system_identifier" && "$database_oid" =~ ^[1-9][0-9]*$ ]]

case "$PG_READY_REQUIRE_SOURCE_IDENTITY" in
  true)
    [[ -n "${PG_READY_DIFFERENT_SYSTEM_IDENTIFIER:-}" ]]
    [[ "${PG_READY_EXPECTED_DATABASE_OID:-}" =~ ^[1-9][0-9]*$ ]]
    [[ "$system_identifier" != "$PG_READY_DIFFERENT_SYSTEM_IDENTIFIER" ]]
    [[ "$database_oid" == "$PG_READY_EXPECTED_DATABASE_OID" ]]
    ;;
  false)
    [[ -z "${PG_READY_DIFFERENT_SYSTEM_IDENTIFIER:-}" ]]
    [[ -z "${PG_READY_EXPECTED_DATABASE_OID:-}" ]]
    ;;
  *) exit 64 ;;
esac

printf 'database-ready|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s\n' \
  "$database_name" "$version" "$checksums" "$encoding" "$provider" "$collation" "$ctype" \
  "$owner" "$system_identifier" "$database_oid" "$vector_version" "$vector_schema"
