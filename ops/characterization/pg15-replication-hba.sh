#!/usr/bin/env bash
set -euo pipefail

: "${PGDATA:?PGDATA is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
printf 'host replication %s all scram-sha-256\n' "$POSTGRES_USER" >>"$PGDATA/pg_hba.conf"
