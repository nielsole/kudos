# CLAUDE.md — Kudos Codebase Guide

This file provides guidance for AI assistants working in this repository.

---

## Project Overview

**Kudos** is a self-hosted "like button" widget. Users hover over the widget for ~1.5 seconds to give kudos. The system consists of:

- A **Go backend** that stores and serves kudos counts using Redis.
- A **vanilla JavaScript/SCSS frontend widget** that embeds in any webpage.
- A **Playwright E2E test suite** that tests the full stack end-to-end.

---

## Repository Structure

```
kudos/
├── *.go                    # Go backend (main.go, handlers.go, utils.go, *_test.go)
├── go.mod / go.sum         # Go module dependencies
├── Dockerfile              # Multi-stage Docker build (Go → scratch)
├── docker-compose.yml      # Local dev: Redis + backend containers
├── Makefile                # Common dev tasks
├── .gitlab-ci.yml          # GitLab CI pipeline
│
├── frontend/               # JavaScript widget + SCSS styles
│   ├── kudosplease.js      # Unminified widget source
│   ├── kudosplease.min.js  # Minified build output (generated)
│   ├── kudosplease.scss    # Widget styles source
│   ├── kudosplease.css     # Compiled CSS (generated)
│   ├── kudosplease.css.map # CSS source map (generated)
│   ├── bundle.sh           # Build script (runs minify + sass)
│   ├── test.html           # Manual test fixture
│   ├── e2e-fixture.html    # E2E test HTML fixture
│   └── package.json        # Build tool dependencies (minify, sass)
│
├── e2e/                    # Playwright end-to-end tests
│   ├── tests/kudos.spec.js # Test suite (8 test cases)
│   ├── playwright.config.js
│   ├── global-setup.js     # Starts Redis + backend before tests
│   ├── global-teardown.js  # Stops processes after tests
│   └── package.json        # Playwright + http-server deps
│
├── docs/                   # Additional documentation
├── static/kudos.png        # Kudos icon asset
└── README.md               # End-user documentation
```

---

## Technology Stack

| Layer       | Technology                                  |
|-------------|---------------------------------------------|
| Backend     | Go 1.15+ (gorilla/mux, go-redis v8)         |
| Database    | Redis (key-value counter store)             |
| Frontend    | Vanilla JavaScript + SCSS (no frameworks)   |
| E2E Tests   | Playwright 1.56.1 (Chromium)                |
| Build       | minify (~8.0.0), sass (~1.52.1)             |
| Container   | Docker (multi-stage, `scratch` base image)  |
| CI/CD       | GitLab CI with Kaniko                       |

---

## Development Workflow

### Backend

```bash
# Build binary
go build -o kudos .

# Run tests (requires Redis on localhost:6379)
go test ./...

# Start Redis for local dev (Docker)
sudo docker run --name kudos-redis -d -p 6379:6379 redis

# Run backend locally
./kudos -port :8080 -redis-address localhost:6379

# Full test cycle (Makefile)
make test
```

### Frontend

```bash
cd frontend
npm install
./bundle.sh   # Compiles kudosplease.scss → .css and minifies .js
```

### E2E Tests

```bash
# Full E2E run (installs deps, launches Chromium, starts Redis + backend)
make e2e-test

# Or manually:
cd e2e
npm install
npx playwright install chromium
npx playwright test

# Headed mode for debugging
npx playwright test --headed

# Debug mode (step through)
npx playwright test --debug
```

### Docker / Full Stack

```bash
# Start everything (Redis + backend)
docker-compose up

# Build image
docker build -t kudos .
```

---

## Key Commands (Makefile)

| Target      | Description                                  |
|-------------|----------------------------------------------|
| `make build`    | Compile Go binary (`go build -o kudos`)  |
| `make test`     | Run Go unit/integration tests with Redis |
| `make e2e-test` | Run Playwright E2E test suite            |

---

## Backend Architecture

### Entry Point (`main.go`)

- Parses flags: `-port` (default `:8080`), `-admin-port` (default `:8081`), `-redis-address`, `-redis-password`, `-redis-db`
- `REDIS_PASSWORD` env var overrides the flag
- Two HTTP servers: main API on `:8080`, admin/metrics on `:8081`

### API Endpoints

**Main API (`main.go`, served on `:8080`)**

| Method | Path             | Description                                  |
|--------|------------------|----------------------------------------------|
| GET    | `/?url=<url>`    | Get kudos count for a URL                    |
| POST   | `/?url=<url>`    | Increment kudos count; returns new count     |
| GET    | `/getclaps`      | Get count using HTTP `Referer` header        |
| POST   | `/update-claps`  | Increment using `Referer` header             |

**Admin API (`:8081`)**

| Method | Path            | Description                                  |
|--------|-----------------|----------------------------------------------|
| GET    | `/metrics`      | Prometheus-style metrics for all URLs        |
| GET    | `/debug/pprof/` | Go pprof profiling endpoints                 |

All API endpoints set `Access-Control-Allow-Origin: *` and no-cache headers.

### Handler Pattern (`handlers.go`)

Handlers use a closure pattern to inject the Redis client:

```go
func handler(fn func(http.ResponseWriter, *http.Request, *redis.Client)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fn(w, r, client)
    }
}
```

### Redis Data Model

- **Keys:** Full URL strings (e.g., `http://example.com/page`)
- **Values:** Integer counters stored as strings
- **Operations:** `INCR` (atomic increment), `GET`, `SCAN` (for domain aggregation)
- No TTL — counts persist indefinitely

### Domain/Subdomain Logic (`utils.go`)

`IsSubdomainOf(subdomain, domain string) bool` — determines if a URL belongs to a domain. Used for aggregating counts across a domain.

---

## Frontend Architecture

### Widget (`frontend/kudosplease.js`)

- **Pattern:** IIFE returning a `KudosPlease` constructor
- **No external dependencies** — zero runtime deps
- **Initialization:**
  ```html
  <div class="kudos" data-url="http://example.com/page" data-amount="5"></div>
  <script>new KudosPlease({ el: '.kudos', duration: 1500, persistent: true, status: {...} });</script>
  ```

### Widget States

| State    | CSS Class  | Meaning                               |
|----------|------------|---------------------------------------|
| `alpha`  | (none/default) | No kudos given yet                |
| `beta`   | (icon class)   | Some kudos exist on this URL      |
| `gamma`  | `finish`       | This user has given kudos         |
| Active   | `active`       | User is currently hovering        |

### Constructor Options

```javascript
new KudosPlease({
    el: '.kudos',       // CSS selector for widget container
    duration: 1500,     // Hover duration in ms to register kudos
    persistent: true,   // Use localStorage to remember user's kudos
    status: {
        alpha: 'icon-alpha-class',
        beta: 'icon-beta-class',
        gamma: 'icon-gamma-class'
    }
});
```

### Backend URL

The widget has a hardcoded backend URL `https://kudos.niels-ole.com/`. For self-hosting, this must be updated in `kudosplease.js` before running `bundle.sh`.

### Event Handling

- Supports both **mouse** (hover) and **touch** (single finger) events
- Touch: single finger only (multi-touch allows native zoom)
- A timer starts on hover/touchstart and is cancelled on leave/touchend
- On timer completion, the count is incremented via POST to backend

### Build Output

Running `bundle.sh` regenerates:
- `kudosplease.min.js` — minified from `kudosplease.js`
- `kudosplease.css` + `kudosplease.css.map` — compiled from `kudosplease.scss`

Commit both source and generated files.

---

## Testing

### Backend Tests (`handlers_test.go`, `utils_test.go`)

- `TestRedisIntegration` — integration test; requires a live Redis on `localhost:6379`
- `TestIsSubdomainOf` — unit tests for subdomain matching (6 cases)
- Run with: `go test ./...`

### E2E Tests (`e2e/tests/kudos.spec.js`)

8 test cases covering:
1. Widget renders with zero count
2. Hover triggers `active` class
3. Releasing hover before timer cancels gesture
4. Full hover increments count and transitions to `finish` state
5. Redis count is incremented on the backend
6. Reload shows `finish` state (localStorage + backend persistence)
7. CSS transform is applied on hovering a finished widget
8. Widget initializes in `finish` state when localStorage is pre-set

**E2E infrastructure:**
- `global-setup.js` starts Redis on `:6380` and the backend on `:8082/:8083`
- `global-teardown.js` terminates those processes
- Playwright intercepts requests to the hardcoded production URL and proxies to `localhost:8082`
- Tests run with `workers: 1` to prevent Redis key collisions

---

## Environment Variables & Ports

| Variable / Flag     | Default          | Description                   |
|---------------------|------------------|-------------------------------|
| `-port`             | `:8080`          | Main API listen address       |
| `-admin-port`       | `:8081`          | Admin/metrics listen address  |
| `-redis-address`    | `localhost:6379` | Redis connection address       |
| `-redis-password`   | `""`             | Redis password (or `REDIS_PASSWORD` env) |
| `-redis-db`         | `0`              | Redis database index           |

**E2E-specific ports:** backend on `:8082/:8083`, Redis on `:6380`, frontend fixture server on `:3000`.

---

## Code Conventions

### Go

- Standard Go formatting enforced — always run `gofmt -w .` before committing (CI will reject non-formatted code)
- Error handling: log with `println`, return HTTP 400 for bad client input; no panics
- Handlers accept `(w http.ResponseWriter, r *http.Request, client *redis.Client)` and are wrapped by the `handler()` closure
- Context with timeout for all Redis operations

### JavaScript

- No transpilation — plain ES5-compatible JavaScript for broad browser support
- No build-time imports; all code in a single IIFE
- Polyfills injected at runtime if feature detection fails (e.g., `classList`)
- Widget state tracked via CSS classes and `data-*` attributes

### SCSS

- Variable-driven design (durations, colors, dimensions)
- Nested selectors; naming is BEM-adjacent
- Shared animation definitions at top of file

### Commits & CI

- `gofmt` is enforced in CI (`format` job); commits with non-standard formatting will fail
- CI stages: `test` → `build` → `promote`
- Docker images are built with Kaniko and tagged with commit SHA; `latest` is promoted only on `master`/tags

---

## Gotchas & Known Issues

1. **Hardcoded backend URL** in `frontend/kudosplease.js` (`https://kudos.niels-ole.com/`). Update for self-hosting before building.
2. **Duplicate Redis clients in go.mod** — both `go-redis v6` and `v8` are listed. The active code uses `v8`; the v6 entry is unused.
3. **`make test` requires `sudo`** to start the Docker Redis container — this is a legacy quirk.
4. **E2E tests run single-worker** (`workers: 1`) — do not change this without adding per-test URL namespacing.
5. **Generated files are committed** — `kudosplease.min.js`, `kudosplease.css`, and `kudosplease.css.map` live in the repo alongside sources.
6. **No authentication** on any endpoint — the API is fully public; rate limiting or auth must be handled at the infrastructure level.
