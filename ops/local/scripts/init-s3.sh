#!/bin/sh
set -eu

root_access_key="$(cat /run/secrets/minio-root-access-key.txt)"
root_secret_key="$(cat /run/secrets/minio-root-secret-key.txt)"
runtime_access_key="$(cat /run/secrets/s3-access-key-id.txt)"
runtime_secret_key="$(cat /run/secrets/s3-secret-access-key.txt)"
attempt=0

until mc alias set local http://minio:9000 "$root_access_key" "$root_secret_key" >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 60 ]; then
    echo "MinIO did not become ready" >&2
    exit 1
  fi
  sleep 1
done

for bucket in \
  harpy-documents \
  harpy-documents-legacy \
  harpy-analytics-dashboard-exports \
  harpy-form-exports \
  harpy-form-import-sources \
  harpy-iam-system-exports \
  harpy-iam-system-import-sources
do
  mc mb --ignore-existing "local/$bucket" >/dev/null
  mc version enable "local/$bucket" >/dev/null
  mc anonymous set none "local/$bucket" >/dev/null
done

mc admin user add local "$runtime_access_key" "$runtime_secret_key" >/dev/null
mc admin policy create local harpy-runtime /scripts/s3-runtime-policy.json >/dev/null
mc admin policy attach local harpy-runtime --user "$runtime_access_key" >/dev/null

echo "Local S3 buckets are versioned and private"
