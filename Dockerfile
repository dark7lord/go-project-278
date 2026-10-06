# 1) Build frontend
FROM node:24-alpine AS frontend-builder
WORKDIR /build/frontend

COPY package*.json ./

# Only the prebuilt dist is needed: no package gets to run install scripts
RUN --mount=type=cache,target=/root/.npm \
  npm ci --prefer-offline --no-audit --ignore-scripts

# 2) Build backend
FROM golang:1.26-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /build/code

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/app .

# 3) Runtime
FROM alpine:3.22

# Everything runs as the app user. No file capabilities: a runtime that drops
# them (Render does) refuses to execute such a binary at all.
RUN apk add --no-cache ca-certificates tzdata bash caddy \
  && adduser -D app

WORKDIR /app

ENV GIN_MODE=release

COPY --from=backend-builder /build/app /app/bin/app
COPY --from=frontend-builder \
  /build/frontend/node_modules/@hexlet/project-url-shortener-frontend/dist \
  /app/public

COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh

COPY Caddyfile /etc/caddy/Caddyfile

USER app

EXPOSE 80

CMD ["/app/bin/run.sh"]