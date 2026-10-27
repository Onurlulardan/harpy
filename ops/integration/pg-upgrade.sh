#!/usr/bin/env bash
set -euo pipefail

old_data=/var/lib/postgresql/15/source
new_data=/var/lib/postgresql/18/docker
work_dir=/var/lib/postgresql/upgrade-work
manifest_path="$old_data/backup_manifest"
evidence_path="${PG_UPGRADE_EVIDENCE_PATH:-}"
expected_manifest_file="${PG_UPGRADE_EXPECTED_MANIFEST_SHA256_FILE:-}"
old_started=0
password_file=
evidence_temp=

cleanup() {
  if [[ "$old_started" == "1" ]]; then
    /usr/lib/postgresql/15/bin/pg_ctl --pgdata="$old_data" \
      --mode=immediate --wait stop >/dev/null 2>&1 || true
  fi
  [[ -z "$password_file" ]] || rm -f "$password_file"
  [[ -z "$evidence_temp" ]] || rm -f "$evidence_temp"
}
trap cleanup EXIT

if [[ "$(id -u)" == "0" ]]; then
  install -d -m 0700 -o postgres -g postgres "$new_data" "$work_dir"
  if [[ -n "$evidence_path" ]]; then
    install -d -m 0700 -o postgres -g postgres "$(dirname "$evidence_path")"
  fi
  exec gosu postgres "$0" "$@"
fi

: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${PG_UPGRADE_USER:?PG_UPGRADE_USER is required}"
[[ "$PG_UPGRADE_USER" =~ ^[a-z_][a-z0-9_]{0,62}$ ]]
[[ "$(cat "$old_data/PG_VERSION")" == "15" ]]
[[ ! -e "$new_data/PG_VERSION" ]]
[[ -f "$manifest_path" ]]

manifest_sha256="sha256:$(sha256sum "$manifest_path" | cut -d' ' -f1)"
/usr/lib/postgresql/15/bin/pg_verifybackup --exit-on-error "$old_data"
if [[ -n "$expected_manifest_file" ]]; then
  [[ "$(cat "$expected_manifest_file")" == "$manifest_sha256" ]]
fi

/usr/lib/postgresql/15/bin/pg_ctl \
  --pgdata="$old_data" \
  --log="$work_dir/pg15-recovery.log" \
  --options="-p 50431 -k /tmp -c listen_addresses=''" \
  --wait start
old_started=1
/usr/lib/postgresql/15/bin/pg_ctl \
  --pgdata="$old_data" --mode=fast --wait stop
old_started=0

/usr/lib/postgresql/15/bin/pg_checksums --enable --pgdata="$old_data"
/usr/lib/postgresql/15/bin/pg_checksums --check --pgdata="$old_data"

password_file=$(mktemp)
chmod 0600 "$password_file"
printf '%s\n' "$POSTGRES_PASSWORD" >"$password_file"
export PGPASSWORD="$POSTGRES_PASSWORD"

/usr/lib/postgresql/18/bin/initdb \
  --pgdata="$new_data" \
  --encoding=UTF8 \
  --locale=en_US.utf8 \
  --data-checksums \
  --username="$PG_UPGRADE_USER" \
  --auth-local=trust \
  --auth-host=scram-sha-256 \
  --pwfile="$password_file"

cd "$work_dir"
upgrade=(
  --old-bindir=/usr/lib/postgresql/15/bin
  --new-bindir=/usr/lib/postgresql/18/bin
  --old-datadir="$old_data"
  --new-datadir="$new_data"
  --old-port=50432
  --new-port=50433
  --socketdir=/tmp
  --username="$PG_UPGRADE_USER"
)

/usr/lib/postgresql/18/bin/pg_upgrade --check "${upgrade[@]}"
/usr/lib/postgresql/18/bin/pg_upgrade --copy --jobs=2 "${upgrade[@]}"
printf '\nhost all all all scram-sha-256\n' >>"$new_data/pg_hba.conf"

evidence=$(printf \
  '{"contract_version":"pg15-pg18-upgrade.v1","manifest_sha256":"%s","result":"succeeded"}' \
  "$manifest_sha256")
if [[ -n "$evidence_path" ]]; then
  [[ ! -e "$evidence_path" ]]
  evidence_temp=$(mktemp "$(dirname "$evidence_path")/.upgrade.XXXXXX")
  printf '%s\n' "$evidence" >"$evidence_temp"
  chmod 0600 "$evidence_temp"
  sync "$evidence_temp"
  ln "$evidence_temp" "$evidence_path"
  rm -f "$evidence_temp"
  evidence_temp=
fi
printf '%s\n' "$evidence"
