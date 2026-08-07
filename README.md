# go-ecommerce-app

A full-stack e-commerce app built to go deep on Go backend fundamentals: clean layering, real payment gateway integration, and a production deployment — not a tutorial clone.

**Live demo:** https://www.go-commerce-project.my.id
**API:** https://api.go-commerce-project.my.id/api

Demo login is available on the login page (customer and admin roles) — no signup required to try it.

## What it does

- Browse products by category, add to cart, check out
- Real payment flow via **Midtrans** (sandbox), with order status driven by payment callbacks
- Auth: email/password with email verification, JWT access tokens + refresh token rotation, password reset
- Admin-only product management (create/update/delete), decoupled from public read access
- Background sweeper that expires stale unpaid orders on an interval

## Stack

**Backend** (`api/`)
- Go, [chi](https://github.com/go-chi/chi) router
- PostgreSQL via [pgx](https://github.com/jackc/pgx) + [sqlc](https://sqlc.dev/) for typed queries
- Redis (token/session state)
- [Midtrans](https://midtrans.com/) for payments, [Resend](https://resend.com/) for transactional email
- JWT auth (golang-jwt), validation via go-playground/validator
- Tests with testify + uber-go/mock

**Frontend** (`client/`)
- React 19 + TypeScript + Vite
- TanStack Query, React Hook Form + Zod, React Router
- Tailwind CSS v4, custom neobrutalist design system

**Deployment**
- API: Docker Compose on a VPS (Postgres + Redis + app containers)
- Frontend: Vercel

## Architecture

The API follows a layered structure: `handler → service → repository`, with `sqlc`-generated queries and interface-based repositories/services for mocking in tests.

```
api/internal/
  handler/     HTTP handlers (chi routes)
  service/     business logic
  repository/  sqlc-generated queries + interfaces
  model/       domain types
  middleware/  auth (JWT), redis-backed checks
  payment/     Midtrans client
  email/       Resend client
  config/      env-based configuration
```

## Running locally

**Backend**
```bash
cd api
cp .env.production.example .env   # fill in DB/Redis/Midtrans/Resend values
docker compose up -d               # postgres + redis
go run ./cmd/api
```

**Frontend**
```bash
cd client
cp .env.example .env.local
npm install
npm run dev
```

## Notes

This is a learning project — feedback on the architecture or code is very welcome.
