#!/usr/bin/env bash
# Checks what `make setup` needs and says how to fix a miss.
set -euo pipefail

fail() { echo "preflight: $*" >&2; exit 1; }

command -v go  >/dev/null || fail "install Go (the version in go.mod)"
command -v npm >/dev/null || fail "install Node.js 20.19+ or 22.12+"
docker info >/dev/null 2>&1           || fail "start Docker"
docker compose version >/dev/null 2>&1 || fail "install Docker Compose v2"

# The port is busy unless it is our own db container holding it.
port="${POSTGRES_PORT:-5432}"
if [ -z "$(docker compose ps -q db)" ] && (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null; then
  fail "port $port is taken: set POSTGRES_PORT in .env and the same port in DATABASE_URL"
fi

echo "preflight: ok"
