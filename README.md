# Link shortener
Shortens links, redirects through them and records every visit; comes with a react-admin dashboard.

### Hexlet tests and linter status:
[![Actions Status](https://github.com/dark7lord/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/dark7lord/go-project-278/actions)
[![ci](https://github.com/dark7lord/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/dark7lord/go-project-278/actions/workflows/ci.yml)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=dark7lord_go-project-278&metric=coverage)](https://sonarcloud.io/summary/new_code?id=dark7lord_go-project-278)
[![Render](https://img.shields.io/badge/Render-deployed-46E3B7?logo=render)](https://go-project-278-ufeg.onrender.com)
<!-- [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=dark7lord_go-project-278&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=dark7lord_go-project-278) -->

## Demo

**Deployed at:** [go-project-278-ufeg.onrender.com](https://go-project-278-ufeg.onrender.com)

## Quickstart

Requirements: Go (the version in `go.mod`), Node.js 20.19+ or 22.12+, Docker with Compose v2.

```bash
git clone https://github.com/dark7lord/go-project-278.git && cd go-project-278
make setup             # checks, .env, tools, dependencies, PostgreSQL, migrations
make dev               # API on :8080, restarted on save + dashboard on :5173
```

Open [localhost:5173](http://localhost:5173): the dashboard proxies `/api` to the API on `:8080`.
`make dev` rebuilds and restarts the API through [air](https://github.com/air-verse/air)
whenever a Go file changes (`.air.toml`); `make run` does the same without restarts.

`make setup` can be re-run at any time, for example after a pull with new migrations.
It first checks the tools above, that Docker is running and that the PostgreSQL port is free,
then does the steps you could also run by hand:

```bash
cp .env.example .env   # only when there is no .env yet
make tools             # golangci-lint binary; sqlc, goose, air built from tools/go.mod
make deps              # Go modules and the frontend package
make db-up             # PostgreSQL 17 in Docker, data kept in the pgdata volume
make db-migrate        # apply migrations
```

If port 5432 is taken, for example by a local PostgreSQL, set `POSTGRES_PORT=5433` in `.env`
and use the same port in `DATABASE_URL`; `make setup` says so when it finds the port busy.

### Docker

The image bundles the API, the built dashboard and Caddy; migrations run on start.
It is what Render builds; there the database is a separate managed PostgreSQL.

```bash
make docker-up         # the image with the compose PostgreSQL, on http://localhost
make docker-down       # stop the image; PostgreSQL and its data stay
```

`compose.yaml` wires the image to the database by the service name `db`.

## Configuration

| Variable          | Required | Default                 | Description                                         |
|-------------------|----------|-------------------------|-----------------------------------------------------|
| `DATABASE_URL`    | yes      |                         | PostgreSQL connection string                        |
| `BASE_URL`        | no       | `http://localhost:8080` | Public origin used to build `short_url`             |
| `REQUEST_TIMEOUT` | no       | `10s`                   | Per-request budget, a positive Go duration          |
| `SENTRY_DSN`      | no       |                         | Sentry DSN; internal 500 errors are reported there  |

`BASE_URL` must be a bare `http(s)` origin: no path, query, fragment or credentials.

## API

The contract is described in [`openapi/openapi.yaml`](openapi/openapi.yaml)
([view in Redoc](https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/dark7lord/go-project-278/master/openapi/openapi.yaml));
`make api-html` builds the same docs locally.

| Method   | Path               | Description                                         |
|----------|--------------------|-----------------------------------------------------|
| `GET`    | `/ping`            | Health check, answers `pong`                        |
| `GET`    | `/r/{code}`        | `302` to the original URL, records a visit          |
| `GET`    | `/api/links`       | A page of links                                     |
| `POST`   | `/api/links`       | Create a link; `short_name` is generated if omitted |
| `GET`    | `/api/links/{id}`  | Get a link                                          |
| `PUT`    | `/api/links/{id}`  | Update a link; `short_name` is generated if omitted |
| `DELETE` | `/api/links/{id}`  | Delete a link and its visits                        |
| `GET`    | `/api/link_visits` | A page of visits                                    |

A link body is `{"original_url": "...", "short_name": "..."}`: `original_url` is required
and must be an `http(s)` URL, `short_name` is 3–32 characters.

### Pagination and sorting

Collections are read by pages. `ra-data-simple-rest` sends both ways of asking for one:

- the `range` query parameter, `[start,end]` inclusive, answered with `200`; it wins over the header,
  and a malformed value answers `400`;
- an RFC 9110 `Range` header whose unit is the collection name, answered with `206`:
  `links=0-9`, `links=10-` (from 10 on) or `links=-20` (the last 20).
  A header with another unit, a malformed value or several ranges is ignored.

Without either the first page is read, as if `links=0-` were sent.

- Every page carries `Content-Range: <collection> <first>-<last>/<total>` and `Accept-Ranges: <collection>`.
- A page holds at most 1000 items: a wider range is cut, and `Content-Range` shows what came back.
- A range past the end, or the empty suffix `-0`, answers `416` with `Content-Range: <collection> */<total>`;
  an empty collection answers `200 []` with `*/0` whatever the range.

`sort=["field","ASC|DESC"]` orders the collection before the page is cut.
Links sort by `id`, `original_url`, `short_name`;
visits by `id`, `link_id`, `created_at`, `ip`, `user_agent`, `referer`, `status`.

### Errors and limits

- Errors are JSON: `{"error": "..."}`, or `{"errors": {"<field>": "..."}}` with `422` for invalid fields.
- `400` for malformed JSON, unknown fields, a bad id, range or sort.
- A request body is limited to 1 MiB.
- A request running longer than `REQUEST_TIMEOUT` answers `503 {"error": "request timeout"}`.
- Every response except the `503` timeout carries a generated `X-Request-ID`;
  the request log line and the Sentry event of a `500` carry the same id.
- Logs are `log/slog` text lines; a `5xx` is logged at `ERROR` with the internal error
  the client never sees.

## Examples

Create a link with a chosen short name:

```bash
curl -i localhost:8080/api/links \
  -H 'Content-Type: application/json' \
  -d '{"original_url": "https://example.com/long/path", "short_name": "example"}'
```

```
HTTP/1.1 201 Created

{"id":1,"original_url":"https://example.com/long/path","short_name":"example","short_url":"http://localhost:8080/r/example"}
```

Without `short_name` the server generates one:

```bash
curl localhost:8080/api/links \
  -H 'Content-Type: application/json' \
  -d '{"original_url": "https://go.dev"}'
```

```
{"id":2,"original_url":"https://go.dev","short_name":"gixi-938028","short_url":"http://localhost:8080/r/gixi-938028"}
```

A taken short name:

```
HTTP/1.1 422 Unprocessable Entity

{"errors":{"short_name":"short name already in use"}}
```

Follow the short link:

```bash
curl -i localhost:8080/r/example
```

```
HTTP/1.1 302 Found
Location: https://example.com/long/path
```

A sorted page of links:

```bash
curl -i -G localhost:8080/api/links \
  --data-urlencode 'range=[0,1]' \
  --data-urlencode 'sort=["id","DESC"]'
```

```
HTTP/1.1 200 OK
Content-Range: links 0-1/3

[{"id":3,"original_url":"https://hexlet.io","short_name":"gacize-235272","short_url":"http://localhost:8080/r/gacize-235272"},{"id":2,"original_url":"https://go.dev","short_name":"gixi-938028","short_url":"http://localhost:8080/r/gixi-938028"}]
```

Visits, with an RFC range header:

```bash
curl -i localhost:8080/api/link_visits -H 'Range: link_visits=0-1'
```

```
HTTP/1.1 206 Partial Content
Content-Range: link_visits 0-0/1

[{"id":1,"link_id":1,"created_at":"2026-09-28T18:36:35.585766+05:00","ip":"::1","user_agent":"curl/8.7.1","status":302}]
```

The last link, with a suffix range:

```bash
curl -i localhost:8080/api/links -H 'Range: links=-1'
```

```
HTTP/1.1 206 Partial Content
Content-Range: links 2-2/3

[{"id":3,"original_url":"https://hexlet.io","short_name":"gacize-235272","short_url":"http://localhost:8080/r/gacize-235272"}]
```

A range past the end:

```bash
curl -i -G localhost:8080/api/links --data-urlencode 'range=[10,19]'
```

```
HTTP/1.1 416 Requested Range Not Satisfiable
Content-Range: links */3

{"error":"range not satisfiable"}
```

## Development

```bash
# development
make help              # every target by section; a bare `make` shows it too
make setup             # after a clone; make preflight runs its checks alone
make dev               # PostgreSQL + API restarted on save + dashboard
make run               # the same without restarts (run-front: the dashboard alone)
make build             # build the API binary to bin/app
make deps              # Go modules and the frontend package
make check             # test + lint + build, what CI checks

# tests
make test              # all tests with -race, writes coverage.out (integration ones need Docker)
make test-unit         # unit tests only, no Docker
make test-integration  # internal/app against a PostgreSQL in testcontainers
make cover             # tests + coverage report (cover-html opens it in a browser)

# code quality
make lint              # golangci-lint, expects 0 issues (lint-fix applies fixes)
make fmt               # format the code

# database
make db-up             # start PostgreSQL (db-down stops it, the data stays)
make db-migrate        # apply migrations (db-rollback undoes the latest)
make db-status         # show migration status (db-redo, db-validate)
make db-reset          # drop all local data and migrate from scratch, asks first
make db-gen            # regenerate db/generated from db/queries (sqlc)

# Docker
make docker-up         # the image with PostgreSQL, as on Render (docker-down, docker-build)

# API docs and tools
make api-lint          # validate openapi/openapi.yaml (api-html opens the docs)
make tools             # golangci-lint binary; sqlc, goose, air from tools/go.mod
```

sqlc, goose and air are pinned in `tools/go.mod` and run through `go tool -modfile=tools/go.mod`,
so their dependencies never mix with the app's. To bump one:
`cd tools && go get -tool <module>@<version> && go mod tidy`.

`sqlc` does not delete stale files: after renaming or removing a file in `db/queries`,
remove its `.sql.go` from `db/generated` before `make db-gen`.

## Architecture

Dependencies point inward: `domain ← application ← adapters ← app`.

| Path                          | Role                                                              |
|-------------------------------|-------------------------------------------------------------------|
| `cmd/api`                     | Entry point: loads `.env` and runs the app                        |
| `internal/domain`             | Value types and their validation (URL, short code)                |
| `internal/application`        | Use cases and the ports they need                                 |
| `internal/adapters/http`      | gin handlers, the JSON contract, range/sort parsing, middleware   |
| `internal/adapters/postgres`  | Repositories over the sqlc-generated code                         |
| `internal/adapters/shortcode` | Random short-code generator (`word-123456`)                       |
| `internal/app`                | Composition root: wiring, server, timeouts, graceful shutdown     |
| `internal/config`             | Configuration from the environment                                |
| `db`                          | Migrations, SQL queries and the generated `db/generated`          |

In production Caddy serves the dashboard and proxies everything else to the API,
so the browser talks to one origin; the Vite dev server does the same for `make run`.
