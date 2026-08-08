# go-ecommerce-app — API

A REST API for an e-commerce platform: auth, product catalog, cart, checkout with real payment processing (Midtrans), and transactional email — written in Go with a layered, test-driven architecture.

Live: `api.go-commerce-project.my.id` · Frontend: `www.go-commerce-project.my.id` ([client](../client))

> This is the backend-specific deep dive. For the full-stack overview (demo login, frontend stack, deployment topology), see the [root README](../README.md).

## Why this exists

Most portfolio CRUD APIs stop at "create/read/update/delete a resource." This one goes a step further into the parts that are actually hard to get right: token rotation with reuse detection, idempotent payment webhooks, stock reservation under concurrent checkouts, and rate limiting per-endpoint — the kind of edge cases that show up once real users start hitting the system.

## Feature highlights

- **Auth**: JWT access tokens + rotating refresh tokens (`HttpOnly` cookie). A replayed/stale refresh token revokes the entire session family instead of just failing silently — mitigates token theft.
- **Email verification & password reset**: token-based flows, sent asynchronously via [Resend](https://resend.com) so the HTTP response never blocks on an outbound email API call.
- **Product catalog**: search, category filter, pagination, soft delete, separate admin views that include inactive products.
- **Cart & checkout**: server-side price snapshotting (client never dictates price), stock validated and reserved atomically inside a DB transaction at checkout.
- **Payments**: [Midtrans](https://midtrans.com) Snap integration; webhook authenticated via HMAC signature verification (not JWT), idempotent against replayed notifications.
- **Order sweeper**: background goroutine expires abandoned `pending_payment` orders past a threshold and releases their reserved stock — no manual cleanup, no stuck inventory.
- **Rate limiting**: Redis-backed, keyed by IP, email, or user ID depending on the endpoint (e.g. login is IP+email, password change is per-user).
- **Graceful shutdown**: in-flight HTTP requests, the order sweeper, and any pending async email sends all drain on `SIGTERM` before the process exits.

## Tech stack

| | |
|---|---|
| Language | Go 1.26 |
| Router | [chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL via [pgx](https://github.com/jackc/pgx) + [sqlc](https://sqlc.dev) (typed SQL, no ORM) |
| Cache / rate limiting | Redis |
| Auth | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) + bcrypt |
| Payments | Midtrans |
| Email | Resend |
| Testing | testify + [go.uber.org/mock](https://github.com/uber-go/mock), [miniredis](https://github.com/alicebob/miniredis) for Redis-dependent middleware |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Deployment | Docker, VPS (see `docker-compose.prod.yml`) |

## Architecture

```
cmd/api          entrypoint: config, DI wiring, router, graceful shutdown
internal/
  handler/       HTTP layer — decode, validate, call service, encode
  service/       business logic, transactions, orchestration
  repository/    sqlc-generated typed queries (internal/repository/queries/*.sql)
  middleware/    auth, role checks, Redis rate limiting
  model/         request/response DTOs + validation tags
  payment/       Midtrans client + webhook signature verification
  email/         Resend client
pkg/
  security/      JWT, password hashing, token hashing
  response/      uniform JSON envelope
  validation/    validator setup + custom rules (e.g. decimal price)
migrations/      versioned SQL, up/down
```

Each handler depends on a service through a concrete struct (not an interface) but each service depends on `repository.Querier`, an interface — that's the seam mocked in tests via `go.uber.org/mock`, so handler and service tests run against fakes with zero database or Redis required.

## Running locally

```bash
docker compose up -d   # postgres (localhost:5432, password secret123) + redis (localhost:6379)

# create .env in this directory — see the table below for required vars.
# DATABASE_URL and REDIS_ADDR/REDIS_PASSWORD must match docker-compose.yml's
# values above; Midtrans/Resend keys can be sandbox/test keys.

make migrate-up
go run ./cmd/api
```

`.env.production.example` documents the same variables for the VPS deploy (`docker-compose.prod.yml`) — its values are for production, not this local setup.

Required env vars (see `internal/config/config.go` for full parsing/defaults):

| Var | Purpose |
|---|---|
| `DATABASE_URL` | Postgres connection string |
| `JWT_SECRET` | signs access tokens |
| `REDIS_ADDR`, `REDIS_PASSWORD` | rate limiting store |
| `MIDTRANS_SERVER_KEY`, `MIDTRANS_CLIENT_KEY`, `MIDTRANS_MERCHANT_ID`, `MIDTRANS_ENV` | payments |
| `RESEND_API_KEY`, `RESEND_FROM_ADDRESS` | transactional email |
| `FRONTEND_URL` | CORS allowlist (exact match, one origin) |

## Testing

```bash
go test ./... -race -cover
```

Handler and service layers are unit-tested against mocked repositories/email senders (`internal/*/mocks`, generated with `mockgen` — see `internal/repository/generate.go`). Async code paths (email dispatch on register/reset/resend) are tested deterministically by waiting on a tracked `sync.WaitGroup` rather than sleeping, so tests are fast and don't race the background goroutine.

CI (`.github/workflows/ci.yml`) runs `gofmt`, `go vet`, `go mod tidy` verification, and the full test suite with the race detector on every push/PR to `master`.

## API reference

See [`docs/API.md`](docs/API.md) for the full endpoint list — auth requirements, rate limits, and request/response shapes.

## What I'd change at real scale

Being upfront about the current limits rather than pretending they don't exist:

- **No distributed locking on stock decrement** beyond the single DB transaction — fine at current scale, would need `SELECT ... FOR UPDATE` tuning or a queue-based reservation system under heavy concurrent checkout load.
- **No OpenAPI/Swagger spec** — `docs/API.md` is hand-maintained; would generate this from code annotations if the API surface grew much further.
- **No structured metrics/tracing** (Prometheus, OpenTelemetry) — currently just structured `slog` logs; fine for a single-VPS deployment, not for diagnosing latency across services.
- **No integration test suite against a real Postgres/Redis** (e.g. via testcontainers) — current tests mock the repository layer, which is fast but doesn't catch SQL-level bugs (bad joins, wrong constraint behavior). Would add a thin integration layer for the checkout/payment path specifically, since that's where correctness matters most.
