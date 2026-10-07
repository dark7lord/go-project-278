#!/usr/bin/env bash
set -euo pipefail

echo "[run.sh] Starting Caddy"
caddy run --config /etc/caddy/Caddyfile &

echo "[run.sh] Applying migrations"
/app/bin/migrate up

echo "[run.sh] Starting Go app"
exec /app/bin/app
