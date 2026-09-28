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

Requirements: Go 1.26, Node.js, PostgreSQL, Docker (for integration tests and the image).

`.env.example` works as is against a local PostgreSQL; to start one in Docker:

```bash
docker run -d --name link-shortener-db -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=link_shortener \
  postgres:17-alpine
```

```bash
cp .env.example .env
make tools             # golangci-lint, sqlc, goose (pinned versions)
make deps              # Go modules and the frontend package
make goose-up          # apply migrations
make run               # API on :8080 + dashboard on :5173
```

Open [localhost:5173](http://localhost:5173): the dashboard proxies `/api` to the API on `:8080`.

### Docker

The image bundles the API, the built dashboard and Caddy; migrations run on start.

```bash
make docker-build
make docker-run        # reads .env, serves on :80
```

Inside the container `localhost` is the container itself: point `DATABASE_URL`
at `host.docker.internal` to reach a database on the host.

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
| `GET`    | `/api/links`       | List links (paginated with a range)                 |
| `POST`   | `/api/links`       | Create a link; `short_name` is generated if omitted |
| `GET`    | `/api/links/{id}`  | Get a link                                          |
| `PUT`    | `/api/links/{id}`  | Update a link; `short_name` is generated if omitted |
| `DELETE` | `/api/links/{id}`  | Delete a link and its visits                        |
| `GET`    | `/api/link_visits` | List visits (paginated with a range)                |

A link body is `{"original_url": "...", "short_name": "..."}`: `original_url` is required
and must be an `http(s)` URL, `short_name` is 3–32 characters.

### Pagination and sorting

Collections take an inclusive range `[start,end]`, as the `range` query parameter
or the `Range` header (the query parameter wins), which is what `ra-data-simple-rest` sends.

- A page answers `200` with `Content-Range: <collection> <first>-<last>/<total>`.
- A range past the end answers `416` with `Content-Range: <collection> */<total>`;
  an empty collection answers `200 []` with `*/0`.
- A page holds at most 1000 items; a wider range answers `400`.
- Without a range the whole collection is returned, with no `Content-Range`.

`sort=["field","ASC|DESC"]` orders a page and requires a range.
Links sort by `id`, `original_url`, `short_name`;
visits by `id`, `link_id`, `created_at`, `ip`, `user_agent`, `referer`, `status`.

### Errors and limits

- Errors are JSON: `{"error": "..."}`, or `{"errors": {"<field>": "..."}}` with `422` for invalid fields.
- `400` for malformed JSON, unknown fields, a bad id, range or sort.
- A request body is limited to 1 MiB.
- A request running longer than `REQUEST_TIMEOUT` answers `503 {"error": "request timeout"}`.
- Every response except the `503` timeout carries a generated `X-Request-ID`;
  the Sentry event of a `500` is tagged with it.

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

Visits, with the range in a header:

```bash
curl -i localhost:8080/api/link_visits -H 'Range: [0,1]'
```

```
HTTP/1.1 200 OK
Content-Range: link_visits 0-0/1

[{"id":1,"link_id":1,"created_at":"2026-09-28T16:18:00.072583+05:00","ip":"::1","user_agent":"curl/8.7.1","status":302}]
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
make test       # run all tests (integration ones need Docker)
make lint       # run linters
make cover      # run tests with coverage report
make check      # test + lint + build + clean
make build      # build binary to bin/app
make clean      # remove build artifacts
```

`make help` lists every target.

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
