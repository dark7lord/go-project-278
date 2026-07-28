#!/usr/bin/env sh
set -eu

echo "[run.sh] Starting service"

echo "[run.sh] Running DB migrations"
goose -dir ./migrations postgres "${DATABASE_URL}" up

echo "[run.sh] Starting Go app"
exec /app/bin/app