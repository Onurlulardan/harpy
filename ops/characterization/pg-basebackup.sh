#!/usr/bin/env bash
set -euo pipefail

backup_data=/var/lib/postgresql/15/backup
evidence_dir=/evidence
manifest_path="$backup_data/backup_manifest"
json_path="$evidence_dir/backup.json"
digest_path="$evidence_dir/backup-manifest.sha256"
temporary=

cleanup() {
  [[ -z "$temporary" ]] || rm -f "$temporary"
}
trap cleanup EXIT

publish() {
  target="$1"
  content="$2"
  [[ ! -e "$target" ]]
  temporary=$(mktemp "$evidence_dir/.backup.XXXXXX")
  printf '%s\n' "$content" >"$temporary"
  chmod 0600 "$temporary"
  sync "$temporary"
  ln "$temporary" "$target"
  rm -f "$temporary"
  temporary=
}

if [[ "$(id -u)" == "0" ]]; then
  install -d -m 0700 -o postgres -g postgres "$backup_data" "$evidence_dir"
  exec gosu postgres "$0" "$@"
fi

: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${PG_BACKUP_USER:?PG_BACKUP_USER is required}"
[[ "$PG_BACKUP_USER" =~ ^[a-z_][a-z0-9_]{0,62}$ ]]
[[ ! -e "$backup_data/PG_VERSION" ]]
[[ -z "$(find "$backup_data" -mindepth 1 -maxdepth 1 -print -quit)" ]]
export PGPASSWORD="$POSTGRES_PASSWORD"
umask 077

pg_basebackup \
  --host=postgres \
  --port=5432 \
  --username="$PG_BACKUP_USER" \
  --pgdata="$backup_data" \
  --format=plain \
  --wal-method=stream \
  --checkpoint=fast
pg_verifybackup --exit-on-error "$backup_data"

manifest_sha256="sha256:$(sha256sum "$manifest_path" | cut -d' ' -f1)"
manifest_bytes="$(wc -c <"$manifest_path" | tr -d ' ')"
evidence=$(printf \
  '{"contract_version":"pg15-physical-backup.v1","manifest_sha256":"%s","manifest_bytes":%s,"result":"verified"}' \
  "$manifest_sha256" "$manifest_bytes")
publish "$digest_path" "$manifest_sha256"
publish "$json_path" "$evidence"
printf '%s\n' "$evidence"
