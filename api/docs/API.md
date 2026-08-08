# API Reference

Base path in production: `https://api.go-commerce-project.my.id/api`
Base path locally: `http://localhost:8080/api`

All responses share one envelope:

```json
// success
{ "success": true, "message": "...", "data": { } }

// error
{ "success": false, "message": "..." }
```

**Auth**: JWT access token via `Authorization: Bearer <token>`, issued by `/users/login`. Refresh tokens are delivered as an `HttpOnly` cookie (`refresh_token`), rotated on every use with reuse detection (a replayed/stale token revokes the whole family, forcing re-login).

**Rate limiting**: Redis-backed, keyed by IP and/or email depending on the endpoint. Limits noted per-route below.

---

## Users — `/users`

| Method | Path | Auth | Rate limit | Body | Notes |
|---|---|---|---|---|---|
| POST | `/register` | – | 3 / hour / IP | `{ full_name, email, password }` | `password` min 12 chars. Sends a verification email asynchronously (does not block the response). |
| POST | `/login` | – | 5 / 15min / IP+email | `{ email, password }` | Fails with 403 if email isn't verified yet. Returns access token in body, refresh token as cookie. |
| POST | `/logout` | – (cookie) | – | – | Revokes the refresh token from the cookie. |
| POST | `/logout-all` | Bearer | – | – | Revokes every refresh token for the user (all devices/sessions). |
| POST | `/refresh` | – (cookie) | – | – | Rotates the refresh token; reuse of a revoked token revokes the whole family. |
| POST | `/verify-email` | – | – | `{ token }` | Consumes the emailed token, marks the account verified. |
| POST | `/resend-verification` | – | 3 / hour / email | `{ email }` | No-op response either way (doesn't leak whether the email exists). |
| POST | `/forgot-password` | – | 3 / hour / email | `{ email }` | Issues a 15-minute reset token, emailed asynchronously. |
| POST | `/reset-password` | – | – | `{ token, password }` | Also revokes all existing sessions for the account. |
| GET | `/me` | Bearer | – | – | Current user profile. |
| POST | `/change-password` | Bearer | 5 / 15min / user | `{ old_password, new_password }` | Revokes all other sessions on success. |
| PATCH | `/{id}/role` | Bearer + `admin` | – | `{ role }` | `role` is `customer` or `admin`. |

## Products — `/products`, `/admin/products`

| Method | Path | Auth | Query / Body | Notes |
|---|---|---|---|---|
| GET | `/products` | – | `search, category, page, limit` | Active products only. |
| GET | `/products/categories` | – | – | Distinct category list (active products). |
| GET | `/products/{id}` | – | – | 404 if inactive or missing. |
| POST | `/products` | Bearer + `admin` | `CreateProductRequest` | `sku` must be unique; `price` is a decimal string. |
| PUT | `/products/{id}` | Bearer + `admin` | `UpdateProductRequest` | Full replace, including `is_active`. |
| DELETE | `/products/{id}` | Bearer + `admin` | – | Soft delete (flips `is_active`). |
| GET | `/admin/products` | Bearer + `admin` | `search, category, page, limit` | Includes inactive products. |
| GET | `/admin/products/{id}` | Bearer + `admin` | – | Includes inactive products. |
| GET | `/admin/products/categories` | Bearer + `admin` | – | Category list across all products. |

`CreateProductRequest` / `UpdateProductRequest`:
```json
{
  "name": "string, 3-255 chars",
  "description": "string, max 2000 chars",
  "price": "decimal string, e.g. \"149000.00\"",
  "stock": "int >= 0",
  "sku": "string, 3-50 chars (create only)",
  "category": "string, max 100 chars",
  "image_url": "url, optional",
  "is_active": "bool (update only)"
}
```

## Cart — `/cart` (all routes require `Bearer`)

| Method | Path | Body | Notes |
|---|---|---|---|
| GET | `/cart` | – | Cart contents + computed totals. |
| DELETE | `/cart` | – | Empties the cart. |
| POST | `/cart/items` | `{ product_id, quantity }` | Adding an existing product increments quantity. |
| PATCH | `/cart/items/{id}` | `{ quantity }` | Sets absolute quantity; rejects if it exceeds stock. |
| DELETE | `/cart/items/{id}` | – | Removes one line item. |

## Orders — `/orders`

| Method | Path | Auth | Notes |
|---|---|---|---|
| POST | `/orders/checkout` | Bearer | Snapshots the cart into an order, reserves stock (decrements immediately), creates a Midtrans Snap transaction, returns `snap_token`. Clears the cart on success. |
| GET | `/orders` | Bearer | Orders for the current user. |
| GET | `/orders/{id}` | Bearer | 404 if it doesn't belong to the caller. |
| POST | `/orders/webhook` | – (Midtrans signature) | Midtrans payment notification callback. Authenticated by verifying `signature_key` against `MIDTRANS_SERVER_KEY`, not JWT. Updates order status; restocks the reserved items if the transaction ends up failed/denied/expired/cancelled. |

A background sweeper (`ORDER_SWEEP_INTERVAL`, default 5m) expires pending orders older than `ORDER_SWEEP_THRESHOLD` (default 30m) and releases their reserved stock.
