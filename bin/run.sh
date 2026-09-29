#!/usr/bin/env bash
set -euo pipefail

echo "[run.sh] Starting Caddy"
caddy run --config /etc/caddy/Caddyfile &

echo "[run.sh] Starting Go app (it applies the embedded migrations first)"
exec /app/bin/app
