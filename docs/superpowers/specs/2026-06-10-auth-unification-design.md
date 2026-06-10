# Auth Unification And Setup Flow Design

## Context

Nostalgia has moved to a unified frontend where public pages and `/admin` live in the same Vue application. The backend still keeps two identity surfaces:

- Public users use `users`, `sessions`, and the Gin `/api` login and refresh endpoints.
- Admin users use `admins`, `roles`, `role_permissions`, `sys_menus`, Redis admin sessions, and gRPC-Gateway `/v1` admin login and refresh endpoints.

For a personal blog where the admin area is owner-only, the legacy admin RBAC model is heavier than the product needs. The new model should keep secure session renewal while removing duplicated user/admin authentication.

## Goals

- Keep the `access token + refresh token` session model.
- Remove the separate admin token and user token systems.
- Use `users.role` as the only authorization source.
- Support exactly two roles: `admin` and `visitor`.
- Make normal registration create only `visitor` users.
- Add a one-time setup page and API for creating the first `admin` user.
- Protect setup with a deployment-time `SETUP_TOKEN`.
- Preserve existing article ownership and public user data.
- Keep the PASETO maker implementation available as a non-default token backend.
- Remove legacy `admins`, `roles`, `role_permissions`, and `sys_menus` dependencies after code no longer reads them.

## Non-Goals

- No multi-role RBAC management UI.
- No dynamic menu permission system.
- No migration of `admins.super` into `users`.
- No OAuth or third-party login work.
- No change to the public article browsing model.

## Target Roles

`users.role` becomes the complete role model:

- `admin`: can access `/admin` and all owner-only content management APIs.
- `visitor`: can use public authenticated features such as comments and replies.

The backend must never trust role values from public registration requests. Public registration always writes `visitor`.

## Token Model

Keep short-lived access tokens and long-lived refresh tokens. Replace the separate public/admin payload split with one payload shape:

```text
UserPayload {
  id
  user_id
  username
  role
  issued_at
  expire_at
}
```

The token maker should expose one user token creation and verification path. Admin-only endpoints become normal authenticated endpoints with an extra role check:

```text
RequireAuth
RequireRole(admin)
```

Refresh tokens continue to be backed by `sessions`, so session blocking, token mismatch detection, user-agent/client-IP tracking, and expiry checks remain available.

Both HTTP surfaces must verify the same token format. Today the Gin `/api` server creates JWT tokens while the gRPC-Gateway `/v1` server creates PASETO admin tokens. During this migration, keep JWT as the unified token format for both public and admin traffic, and change gRPC-Gateway admin handlers to verify the same user JWT payload. This preserves compatibility for existing public user sessions and avoids moving all admin content APIs to Gin in the same step.

Keep the PASETO maker implementation in the `token` package, but remove it from the default admin login path. It can stay as a future optional backend or testable implementation as long as it does not reintroduce separate admin payloads.

## JWT Library Upgrade

The current dependency is `github.com/dgrijalva/jwt-go v3.2.0+incompatible`. That repository is archived and the Go vulnerability database reports GO-2020-0017 against `dgrijalva/jwt-go`, with no known fixed version for the v3 module line.

Replace it with the maintained module:

```text
github.com/golang-jwt/jwt/v5 v5.3.1
```

Implementation notes:

- Use v5 APIs and `RegisteredClaims` where useful, while keeping the existing payload JSON contract stable for frontend compatibility.
- Pin validation to the expected HMAC signing method. Do not accept algorithm values merely because they are present in the token header.
- Keep the current minimum symmetric key size check.
- Avoid `ParseUnverified` in request authentication paths.
- Keep existing JWT tests for wrong algorithm, expired token, malformed token, and valid token behavior.
- Add a dependency verification step with `govulncheck ./...` during implementation.

The Go vulnerability database also reports GO-2025-3553 for `golang-jwt/jwt/v5` before `v5.2.2`; `v5.3.1` is above the patched threshold.

## Setup API

Add two public setup endpoints under the Gin API surface:

```text
GET  /api/setup/status
POST /api/setup/admin
```

`GET /api/setup/status` returns whether setup is available:

```json
{
  "initialized": true,
  "setup_available": false
}
```

`POST /api/setup/admin` creates the first admin user only when all conditions are true:

- No user with `role = 'admin'` exists.
- The request includes the correct setup token.
- Username, password, email, and full name pass the existing validation rules.
- Username and email are unique.

Request body:

```json
{
  "setup_token": "deployment-time-token",
  "username": "owner",
  "password": "strong-password",
  "full_name": "Owner",
  "email": "owner@example.com"
}
```

Response body should reuse the existing safe user response and must not return the hashed password or setup token.

After the first admin exists, `POST /api/setup/admin` always returns a conflict-style response and creates no user, even if the setup token remains configured.

## Setup Token

Add `SETUP_TOKEN` to configuration and `.env.example`.

Rules:

- `SETUP_TOKEN` is required only while setup is available.
- The token must be compared on the server.
- It must never be logged.
- It must not be returned by status APIs.
- A missing token should make setup unavailable for admin creation and return a clear server-side configuration error from `POST /api/setup/admin`.

This prevents a public deployment from allowing the first random visitor to claim the admin account.

Existing default-user configuration must not create an admin implicitly. If the current default-user bootstrap remains in the application, it should create only a `visitor` unless it is removed in favor of the setup flow.

## Frontend Setup Flow

Add a `/setup` route in `web/frontend`.

Behavior:

- On load, call `GET /api/setup/status`.
- If initialized, redirect to the public home page or show a short initialized state with a link to login.
- If setup is available, show a form for setup token, username, password, full name, and email.
- On success, route to `/admin/login`.
- On conflict, refresh setup status and show that the system is already initialized.

The page should reuse the current glass UI direction and avoid public copy that suggests the blog is a private profile.

## Admin Login Flow

The visible `/admin/login` route can remain as the admin-facing login page, but it should call the unified user login endpoint. After login, the frontend checks that the returned user has `role = 'admin'`.

If a `visitor` signs in from `/admin/login`, the frontend should immediately clear the token state and show a permission error. The backend must also enforce `role = 'admin'` on every admin API; frontend checks are only UX.

## Backend Authorization

Create a shared authorization path for Gin and gRPC-Gateway-backed handlers:

- `authMiddleware` verifies the unified access token and stores `UserPayload`.
- `requireRole(admin)` rejects missing or non-admin roles.
- Public authenticated routes use only `authMiddleware`.
- Admin management routes use both `authMiddleware` and `requireRole(admin)`.

Existing gRPC admin methods can be migrated in place, but they should no longer call `VerifyAdminToken` or read `admins.role_id`.

Admin content APIs should remain under the current `/v1` gRPC-Gateway namespace during this migration. Only admin authentication and authorization change. Moving admin content APIs from `/v1` to `/api/admin` can be a later cleanup after the identity model is stable.

## Database Migration Strategy

Use additive and cleanup migrations to reduce risk.

### Migration 1: Normalize Users

- Ensure every existing user has a valid role.
- Convert every role outside `admin` and `visitor` to `visitor`.
- Add a database check constraint for `role IN ('admin', 'visitor')`.
- Add an index on `users(role)` for setup status and role checks.
- Do not add `users.is_active` in this migration.

Existing deployments need a clear operator step before or immediately after applying the normalization migration:

```sql
-- Set owner_user_id to the existing owner's users.id before running.
UPDATE users SET role = 'admin' WHERE id = :'owner_user_id';
```

This is only for already deployed databases with an existing real owner user. New deployments use `/setup`.

### Migration 2: Remove Legacy Admin RBAC Tables

After code no longer depends on them, drop:

- `role_permissions`
- `sys_menus`
- `admins`
- `roles`

Drop dependent sqlc queries and generated code in the same implementation phase.

Do not migrate `admins.super`. The old seeded admin is a legacy bootstrap account, not a user profile with article ownership.

## Data Compatibility

Article ownership already references `users.id`, so published content and authorship should remain intact.

Comments already reference `users.id`, so public discussion data remains intact.

Sessions reference `users.id`, so existing public refresh-token sessions can continue to work after the token maker is unified on JWT.

Admin Redis sessions are disposable and can expire naturally or be ignored after the admin token path is removed.

Because JWT remains the unified token format, existing public sessions should remain valid. Legacy admin PASETO sessions will stop being useful after the admin token path is removed; admins should log in again through `/admin/login`.

## API Cleanup

Remove or replace these surfaces:

- `/v1/admin/login`
- `/v1/admin/renew_access`
- Admin-specific token payload creation and verification.
- Admin-specific frontend auth store and refresh logic.
- Menu initialization based on `role_permissions`.

Keep or unify these surfaces:

- `/api/users/login`
- `/api/tokens/renew_access`
- `/api/setup/status`
- `/api/setup/admin`

Admin content APIs may remain under their current route namespace, but authorization should come from unified user auth.

## Frontend State Cleanup

The frontend should keep one auth store that owns:

- access token
- access token expiry
- refresh token
- refresh token expiry
- current user
- role

Admin pages use the same store and route guards:

- Missing token: redirect to `/admin/login`.
- Token valid but role is not `admin`: clear or reject admin navigation and show permission feedback.
- Expired access token with valid refresh token: refresh once through the unified refresh endpoint.

## Testing Plan

Backend:

- Setup status returns uninitialized when no admin exists.
- Setup admin rejects missing or wrong setup token.
- Setup admin creates an `admin` only once.
- Setup admin returns conflict after an admin exists.
- Public registration always creates `visitor`.
- Unified login returns user role.
- Admin-only endpoint rejects unauthenticated users.
- Admin-only endpoint rejects `visitor`.
- Admin-only endpoint accepts `admin`.
- Refresh token still checks session ID, refresh token string, blocked status, and expiry.
- Migrations apply up and down on a test database.

Frontend:

- `/setup` redirects or blocks after initialization.
- `/setup` handles successful first admin creation.
- `/admin/login` rejects visitor users.
- Admin route guard accepts admin users.
- Auth refresh still works for public and admin pages.

Verification:

- Run `make sqlc` after query changes.
- Run `make test` after backend changes.
- Run `cd web/frontend && bun run type-check && bun run build` after frontend changes.

## Rollout Plan

Implement as separate commits:

1. Add setup status/admin API and config.
2. Unify token payload and backend auth middleware.
3. Move admin login and admin route guards to unified auth.
4. Add users role migration and sqlc updates.
5. Remove legacy admin RBAC tables and code.
6. Update docs and deployment notes.

Before deploying to an existing database, choose the existing owner user and promote it to `admin`. For a fresh deployment, configure `SETUP_TOKEN`, open `/setup`, create the first admin, then keep or rotate the setup token according to deployment policy.
