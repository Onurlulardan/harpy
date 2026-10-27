#!/bin/sh
set -eu

fail() {
  printf '%s\n' "v1-oracle preflight: $1" >&2
  exit 64
}

require_digest() {
  name="$1"
  value="$2"
  case "$value" in
    *@sha256:*) ;;
    *) fail "$name must be an OCI reference pinned by sha256 digest" ;;
  esac
}

require_password_policy() {
  name="$1"
  value="$2"
  [ "${#value}" -ge 12 ] || fail "$name is too short"
  case "$value" in *[A-Z]*) ;; *) fail "$name needs an uppercase character" ;; esac
  case "$value" in *[a-z]*) ;; *) fail "$name needs a lowercase character" ;; esac
  case "$value" in *[0-9]*) ;; *) fail "$name needs a number" ;; esac
  case "$value" in *[!A-Za-z0-9]*) ;; *) fail "$name needs a symbol" ;; esac
}

project_kind=
case "${ORACLE_PROJECT_NAME:-}" in
  harpy-v1-oracle | harpy-v1-oracle-*) project_kind=oracle ;;
  harpy-v1-pg18-gate-*) project_kind=pg18-gate ;;
  *) fail "COMPOSE_PROJECT_NAME must use an approved oracle prefix" ;;
esac

require_digest V1_ORACLE_GO_IMAGE "${V1_ORACLE_GO_IMAGE:-}"
require_digest CHARACTERIZER_GO_IMAGE "${CHARACTERIZER_GO_IMAGE:-}"
require_digest V1_ORACLE_DISTROLESS_IMAGE "${V1_ORACLE_DISTROLESS_IMAGE:-}"
require_digest V1_ORACLE_POSTGRES_IMAGE "${V1_ORACLE_POSTGRES_IMAGE:-}"
require_digest V1_ORACLE_PGBOUNCER_IMAGE "${V1_ORACLE_PGBOUNCER_IMAGE:-}"
require_digest V1_ORACLE_REDIS_IMAGE "${V1_ORACLE_REDIS_IMAGE:-}"
require_digest V1_ORACLE_RABBITMQ_BASE_IMAGE "${V1_ORACLE_RABBITMQ_BASE_IMAGE:-}"
require_digest V1_ORACLE_MINIO_IMAGE "${V1_ORACLE_MINIO_IMAGE:-}"
if [ "$project_kind" = pg18-gate ]; then
  require_digest PG18_GATE_POSTGRES_IMAGE "${PG18_GATE_POSTGRES_IMAGE:-}"
fi
printf '%s' "${V1_ORACLE_GATEWAY_IMAGE_DIGEST:-}" | grep -Eq '^sha256:[a-f0-9]{64}$' ||
  fail "V1_ORACLE_GATEWAY_IMAGE_DIGEST must be an exact sha256 image ID"

expected_source_postgres='pgvector/pgvector:pg15@sha256:3073dc147f5b0ca05b36e10e04e73ca180b1e726e453c72a6fcc85f788afd0d1'
expected_target_postgres='pgvector/pgvector:0.8.3-pg18-bookworm@sha256:c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84'
[ "${V1_ORACLE_POSTGRES_IMAGE:-}" = "$expected_source_postgres" ] ||
  fail "V1_ORACLE_POSTGRES_IMAGE must equal the attested PG15 source"
if [ "$project_kind" = pg18-gate ]; then
  [ "${PG18_GATE_POSTGRES_IMAGE:-}" = "$expected_target_postgres" ] ||
    fail "PG18_GATE_POSTGRES_IMAGE must equal the attested PG18 target"
fi
case "${V1_ORACLE_GO_IMAGE:-}" in
  *golang:1.25*@sha256:*) ;;
  *) fail "V1_ORACLE_GO_IMAGE must pin Go 1.25" ;;
esac
case "${CHARACTERIZER_GO_IMAGE:-}" in
  *golang:1.26.5*@sha256:*) ;;
  *) fail "CHARACTERIZER_GO_IMAGE must pin Go 1.26.5" ;;
esac

printf '%s' "${V1_ORACLE_SOURCE_VERSION:-}" | grep -Eq '^[a-f0-9]{40,64}$' ||
  fail "V1_ORACLE_SOURCE_VERSION must be a lowercase source hash"
case "${V1_ORACLE_DELAYED_PLUGIN_URL:-}" in
  https://*) ;;
  *) fail "V1_ORACLE_DELAYED_PLUGIN_URL must use https" ;;
esac
printf '%s' "${V1_ORACLE_DELAYED_PLUGIN_SHA256:-}" | grep -Eq '^[a-f0-9]{64}$' ||
  fail "V1_ORACLE_DELAYED_PLUGIN_SHA256 must be a lowercase sha256"

db_password="${DB_PASSWORD:-}"
rabbitmq_password="${RABBITMQ_PASSWORD:-}"
minio_password="${MINIO_ROOT_PASSWORD:-}"
jwt_secret="${JWT_SECRET:-}"
admin_password="${SUPER_ADMIN_PASSWORD:-}"
iam_user_password="${CHARACTERIZATION_IAM_USER_PASSWORD:-}"
tenant_b_admin_password="${CHARACTERIZATION_TENANT_B_ADMIN_PASSWORD:-}"

[ -n "$db_password" ] || fail "DB_PASSWORD is required"
[ "$db_password" = "${POSTGRES_PASSWORD:-}" ] ||
  fail "DB_PASSWORD and POSTGRES_PASSWORD must match"
[ -n "${REDIS_PASSWORD:-}" ] || fail "REDIS_PASSWORD is required"
[ -n "$rabbitmq_password" ] || fail "RABBITMQ_PASSWORD is required"
[ "$rabbitmq_password" = "${RABBITMQ_DEFAULT_PASS:-}" ] ||
  fail "RABBITMQ_PASSWORD and RABBITMQ_DEFAULT_PASS must match"
[ "${#minio_password}" -ge 12 ] || fail "MINIO_ROOT_PASSWORD is too short"
[ "${#jwt_secret}" -ge 32 ] || fail "JWT_SECRET is too short"
require_password_policy SUPER_ADMIN_PASSWORD "$admin_password"
require_password_policy CHARACTERIZATION_IAM_USER_PASSWORD "$iam_user_password"
require_password_policy CHARACTERIZATION_TENANT_B_ADMIN_PASSWORD "$tenant_b_admin_password"
[ -n "${HARPY_LICENSE_KEY:-}" ] || fail "HARPY_LICENSE_KEY is required"
