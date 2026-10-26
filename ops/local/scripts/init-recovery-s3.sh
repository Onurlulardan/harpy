#!/bin/sh
set -eu

. /run/secrets/recovery-s3-users.env

documents_root_access="$(cat /run/secrets/minio-root-access-key.txt)"
documents_root_secret="$(cat /run/secrets/minio-root-secret-key.txt)"
recovery_root_access="$(cat /run/secrets/recovery-minio-root-access-key.txt)"
recovery_root_secret="$(cat /run/secrets/recovery-minio-root-secret-key.txt)"

wait_for_minio() {
  alias_name="$1"
  endpoint="$2"
  access_key="$3"
  secret_key="$4"
  attempt=0
  until mc alias set "$alias_name" "$endpoint" "$access_key" "$secret_key" >/dev/null 2>&1; do
    attempt=$((attempt + 1))
    if [ "$attempt" -ge 60 ]; then
      echo "Recovery storage dependency did not become ready" >&2
      exit 1
    fi
    sleep 1
  done
}

install_user() {
  alias_name="$1"
  access_key="$2"
  secret_key="$3"
  policy_name="$4"
  policy_file="$5"
  mc admin user add "$alias_name" "$access_key" "$secret_key" >/dev/null
  mc admin policy create "$alias_name" "$policy_name" "$policy_file" >/dev/null
  mc admin policy attach "$alias_name" "$policy_name" --user "$access_key" >/dev/null
}

wait_for_minio documents http://minio:9000 \
  "$documents_root_access" "$documents_root_secret"
wait_for_minio recovery http://recovery-minio:9000 \
  "$recovery_root_access" "$recovery_root_secret"

if ! mc stat recovery/harpy-recovery >/dev/null 2>&1; then
  mc mb --with-lock recovery/harpy-recovery >/dev/null
fi
mc version enable recovery/harpy-recovery >/dev/null
mc anonymous set none recovery/harpy-recovery >/dev/null
mc retention set --default COMPLIANCE 30d recovery/harpy-recovery >/dev/null
mc retention info recovery/harpy-recovery >/dev/null

mc mb --ignore-existing documents/harpy-documents-rehydrate >/dev/null
mc version enable documents/harpy-documents-rehydrate >/dev/null
mc anonymous set none documents/harpy-documents-rehydrate >/dev/null

install_user recovery "$WRITE_ACCESS_KEY_ID" "$WRITE_SECRET_ACCESS_KEY" \
  harpy-recovery-write /scripts/recovery-write-policy.json
install_user recovery "$RETENTION_DELETE_ACCESS_KEY_ID" \
  "$RETENTION_DELETE_SECRET_ACCESS_KEY" harpy-recovery-retention-delete \
  /scripts/recovery-retention-delete-policy.json
install_user recovery "$RESTORE_READ_ACCESS_KEY_ID" "$RESTORE_READ_SECRET_ACCESS_KEY" \
  harpy-recovery-restore-read /scripts/recovery-restore-read-policy.json
install_user documents "$DOCUMENTS_SOURCE_READ_ACCESS_KEY_ID" \
  "$DOCUMENTS_SOURCE_READ_SECRET_ACCESS_KEY" harpy-recovery-documents-source-read \
  /scripts/recovery-documents-source-read-policy.json
install_user documents "$DOCUMENTS_REHYDRATE_WRITE_ACCESS_KEY_ID" \
  "$DOCUMENTS_REHYDRATE_WRITE_SECRET_ACCESS_KEY" \
  harpy-recovery-documents-rehydrate-write \
  /scripts/recovery-documents-rehydrate-write-policy.json

echo "Local recovery storage is private, versioned and COMPLIANCE locked"
