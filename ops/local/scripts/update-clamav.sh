#!/bin/sh
set -eu

attempt=1
max_attempts=12

while ! freshclam --foreground --stdout; do
  if [ "$attempt" -ge "$max_attempts" ]; then
    echo "ClamAV signature update failed after $attempt attempts" >&2
    exit 1
  fi

  delay=$((attempt * 5))
  echo "ClamAV signature update failed; retrying in ${delay}s" >&2
  sleep "$delay"
  attempt=$((attempt + 1))
done
