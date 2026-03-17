#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SQLITE_SOURCE="${SQLITE_SOURCE:-$ROOT_DIR/data/drama_generator.db}"
SQLITE_MARKER="${SQLITE_MARKER:-$ROOT_DIR/data/.sqlite_to_mysql_done}"

cd "$ROOT_DIR"

if [[ ! -f "$SQLITE_SOURCE" ]]; then
  echo "[start_mysql] sqlite source not found: $SQLITE_SOURCE"
  exit 1
fi

echo "[start_mysql] starting mysql container..."
docker compose up -d mysql

echo "[start_mysql] waiting for mysql to become healthy..."
for i in $(seq 1 60); do
  status="$(docker inspect --format='{{.State.Health.Status}}' xinggen-mysql 2>/dev/null || true)"
  if [[ "$status" == "healthy" ]]; then
    break
  fi
  if [[ "$status" == "unhealthy" ]]; then
    echo "[start_mysql] mysql is unhealthy, printing logs:"
    docker logs --tail 120 xinggen-mysql || true
    exit 1
  fi
  sleep 2
done

if [[ "$(docker inspect --format='{{.State.Health.Status}}' xinggen-mysql 2>/dev/null || true)" != "healthy" ]]; then
  echo "[start_mysql] mysql did not become healthy in time"
  docker logs --tail 120 xinggen-mysql || true
  exit 1
fi

echo "[start_mysql] building app image (if needed)..."
docker compose build xinggen-drama

echo "[start_mysql] running sqlite -> mysql migration..."
docker compose run --rm -T \
  --no-deps \
  -v "$ROOT_DIR/data:/legacy-data" \
  xinggen-drama \
  /app/sqlite-to-mysql \
  -source "/legacy-data/$(basename "$SQLITE_SOURCE")" \
  -marker "/legacy-data/$(basename "$SQLITE_MARKER")" \
  -wait-attempts 60 \
  -wait-interval 2s

echo "[start_mysql] starting app container..."
docker compose up -d xinggen-drama

echo "[start_mysql] done."
echo "[start_mysql] API: http://localhost:5678"
