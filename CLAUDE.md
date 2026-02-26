# CLAUDE.md — AI Assistant Guide for Kudos

## Project Overview

**Kudos** is a self-hosted "like button" service. It consists of a lightweight Go REST API backend backed by Redis, and a vanilla JavaScript frontend widget. Users embed the widget on their websites; the widget calls the API to persist and retrieve kudos counts.

- **Backend:** Go 1.15+, Gorilla Mux, Redis, Prometheus metrics
- **Frontend:** Vanilla JS, SCSS (compiled to CSS)
- **Infrastructure:** Docker, Docker Compose, GitLab CI/CD

---

## Repository Structure

```
kudos/
├── main.go              # Entry point: flags, Redis init, HTTP server setup
├── handlers.go          # All HTTP handler functions
├── utils.go             # Subdomain matching utility
├── handlers_test.go     # Integration tests (requires Redis)
├── utils_test.go        # Unit tests for utilities
├── go.mod / go.sum      # Go module and dependencies
├── Makefile             # Build and test targets
├── Dockerfile           # Multi-stage build (golang:1.17.3 → scratch)
├── docker-compose.yml   # Redis + dev/prod service definitions
├── .gitlab-ci.yml       # CI pipeline: test → build → promote
├── frontend/            # Frontend widget source
│   ├── kudosplease.js   # Main widget JS (vanilla, no runtime deps)
│   ├── kudosplease.scss # Widget styles (compiled to kudosplease.css)
│   ├── kudosplease.min.js  # Minified output (committed)
│   ├── kudosplease.css     # Compiled CSS output (committed)
│   ├── bundle.sh        # Build script (minify JS, compile SCSS)
│   ├── package.json     # Dev deps: sass, minify
│   └── test.html        # Manual test page
├── docs/                # Landing page / demo site
│   ├── index.html       # Demo page with embedded widget
│   ├── kudosplease.css  # CSS copy for docs
│   └── kudosplease.min.js  # JS copy for docs
└── static/
    └── kudos.png        # Logo asset
```

---

## Development Workflows

### Running Locally

**Prerequisites:** Go 1.17+, Docker (for Redis)

```bash
# Start Redis
docker run -d -p 6379:6379 redis:alpine

# Build and run
make build
./kudos -port :8080 -admin-port :8081 -redis-address localhost:6379
```

**With Docker Compose** (starts Redis + dev service on :8081):
```bash
docker-compose up
```

### Building

```bash
make build          # Compiles ./kudos binary
docker build -t kudos .  # Docker multi-stage build
```

### Running Tests

```bash
make test           # Spins up Redis via Docker, runs go test
go test ./...       # Requires Redis at 127.0.0.1:6379
```

Tests are in two files:
- `utils_test.go` — pure unit tests, no external dependencies
- `handlers_test.go` — integration tests, require a live Redis instance

### Frontend Build

```bash
cd frontend
npm install
npm test            # Runs bundle.sh (minify JS + compile SCSS)
```

After building, copy updated files to `docs/` if updating the demo site:
```bash
cp frontend/kudosplease.min.js docs/kudosplease.min.js
cp frontend/kudosplease.css docs/kudosplease.css
```

---

## Key Conventions

### Go Code

- **Error handling:** Always check and handle errors; log fatal on startup failures.
- **Context timeouts:** All Redis calls use `context.WithTimeout` (typically 5s). Maintain this pattern on new handlers.
- **CORS:** All endpoints return `Access-Control-Allow-Origin: *`. Keep this on any new endpoints.
- **Flags:** Runtime configuration is via `flag` package (see `main.go`). Add new config options as flags with sensible defaults.
- **Formatting:** Code must pass `gofmt` — the CI pipeline enforces this. Run `gofmt -l .` before committing.

### API Design

- **Main port** (default `:8080`): public-facing API endpoints.
- **Admin port** (default `:8081`): Prometheus metrics (`/metrics`) and pprof (`/debug/pprof/`). Never expose this publicly.
- **Request parameters:** URL-based queries use the `url` query param (URL-encoded). Domain-wide queries use the `domain` param.
- **Legacy endpoints** (`/getclaps`, `/update-claps`): Use the `Referrer` HTTP header instead of query params. Kept for backwards compatibility — do not remove.

### Redis Data Model

- **Keys:** Raw URL strings (e.g., `https://example.com/post/1`)
- **Values:** String integers (INCR / GET)
- **No TTL** is set; data persists indefinitely
- **Domain queries:** Use `SCAN` with a glob pattern, then filter using `isSubdomainOf`
- **No namespacing:** Keys are bare URLs — be careful not to introduce key collisions

### Frontend Widget

- The widget (`kudosplease.js`) is **dependency-free** at runtime. Keep it that way.
- State is persisted in **localStorage** (key: the page URL). Do not change the storage key format without considering migration.
- The default API endpoint is hardcoded as `https://kudos.niels-ole.com/`. Users can override by passing `data-url` or `data-domain` attributes.
- The widget supports both **mouse** and **touch** events. Any interaction changes must preserve both.
- Hover duration (default 1.5s) triggers the kudos increment — this is intentional UX.
- **Compiled outputs are committed**: `kudosplease.min.js`, `kudosplease.css`, and their copies in `docs/` are checked into git. Always rebuild and commit them when changing source files.

---

## API Reference

### Main API (`:8080`)

| Method | Path | Query Params | Description |
|--------|------|-------------|-------------|
| `GET`  | `/`  | `url=<encoded>` | Get kudos count for a specific URL |
| `GET`  | `/`  | `domain=<domain>` | Get total kudos count for a domain (incl. subdomains) |
| `POST` | `/`  | `url=<encoded>` | Increment kudos for a URL |
| `GET`  | `/getclaps` | — | Get claps count (uses `Referer` header) |
| `POST` | `/update-claps` | — | Increment claps (uses `Referer` header) |

**Responses:** Plain text integer (e.g., `42`).

### Admin API (`:8081`)

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/metrics` | Prometheus-format metrics (kudos counts as gauges) |
| `*`    | `/debug/pprof/` | Go pprof profiling endpoints |

---

## Configuration Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `:8080` | Main HTTP server listen address |
| `-admin-port` | `:8081` | Admin/metrics server listen address |
| `-redis-address` | `localhost:6379` | Redis server address |
| `-redis-password` | `""` | Redis AUTH password (also `REDIS_PASSWORD` env var) |
| `-redis-db` | `0` | Redis database index |

---

## CI/CD Pipeline (GitLab)

Defined in `.gitlab-ci.yml`, three stages:

1. **test** — Runs `gofmt -l .` to verify formatting. Fails if any file is not properly formatted.
2. **build** — Builds a Docker image using Kaniko and pushes to the GitLab container registry.
3. **promote** — On `master` branch or tags: re-tags the image as `latest` (and with the git tag if applicable).

**To pass CI:** Ensure `gofmt` is clean before pushing. Run `gofmt -w .` to auto-format.

---

## Common Pitfalls

1. **Don't expose the admin port publicly** — it has no authentication and exposes internal metrics and profiling.
2. **Redis SCAN is O(N)** — the domain count endpoint scans all keys. On large datasets this can be slow; be mindful when testing with many keys.
3. **Compiled frontend files must be updated** — if you change `kudosplease.js` or `kudosplease.scss`, rebuild and commit the compiled outputs.
4. **Integration tests need Redis** — `handlers_test.go` will fail if Redis is not running at `127.0.0.1:6379`. Use `make test` which handles this automatically.
5. **gofmt is enforced by CI** — the pipeline will fail on improperly formatted Go code. Always run `gofmt -w .` before committing.

---

## Dependencies

### Go (go.mod)

- `github.com/gorilla/mux v1.6.0` — HTTP router
- `github.com/go-redis/redis/v8 v8.3.1` — Redis client
- `github.com/go-redis/redis v6.15.9` — Legacy Redis client (kept for compatibility)
- `gopkg.in/redis.v5 v5.2.10` — Legacy Redis client

### Frontend (package.json dev deps)

- `sass ~1.52.1` — SCSS compiler
- `minify ~8.0.0` — JavaScript minifier
