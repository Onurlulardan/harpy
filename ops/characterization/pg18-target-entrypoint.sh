#!/usr/bin/env bash
set -euo pipefail

: "${PGDATA:?PGDATA is required}"
[[ -f "$PGDATA/PG_VERSION" ]]
[[ "$(cat "$PGDATA/PG_VERSION")" == "18" ]]
exec /usr/local/bin/docker-entrypoint.sh "$@"
