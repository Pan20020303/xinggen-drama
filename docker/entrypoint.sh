#!/bin/sh
set -eu

SQLITE_MIGRATION_SOURCE="${SQLITE_MIGRATION_SOURCE:-/app/data/drama_generator.db}"
SQLITE_MIGRATION_MARKER="${SQLITE_MIGRATION_MARKER:-/app/data/.sqlite_to_mysql_done}"

if [ -f "$SQLITE_MIGRATION_SOURCE" ] && [ ! -f "$SQLITE_MIGRATION_MARKER" ]; then
  echo "[entrypoint] detected legacy sqlite database at $SQLITE_MIGRATION_SOURCE"
  echo "[entrypoint] starting sqlite -> mysql migration"
  /app/sqlite-to-mysql \
    -source "$SQLITE_MIGRATION_SOURCE" \
    -marker "$SQLITE_MIGRATION_MARKER" \
    -wait-attempts 60 \
    -wait-interval 2s
  echo "[entrypoint] sqlite -> mysql migration finished"
fi

exec "$@"
