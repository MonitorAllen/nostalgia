# Admin User Management Design

## Goal

Add a practical `/backend` user management module for the site owner to manage public `visitor` accounts. The first version focuses on daily operations: finding users, editing basic profile fields, and disabling or restoring accounts. It intentionally avoids admin role management, backend user creation, and password reset so the module stays safe and aligned with the current single-admin architecture.

## Scope

In scope:

- Add a backend "用户" navigation item and `/backend/users` page.
- List all `visitor` users by default, including enabled and disabled accounts.
- Search by username, full name, or email.
- Filter by status: all, enabled, disabled.
- Paginate with page number navigation and adjustable page size.
- Edit visitor profile fields: full name, email, and email verification state.
- Disable visitor accounts while keeping historical content visible.
- Restore disabled visitor accounts.
- Block the disabled user's existing sessions so refresh attempts fail.
- Prevent disabled users from logging in again.

Out of scope:

- Creating users in the backend. Users continue to register from the public frontend.
- Resetting passwords. A public "forgot password" flow will be designed separately.
- Changing roles or creating more admins.
- Managing the current admin account from this module.
- Audit logs, login history, and session detail views.

## Existing Context

Nostalgia now uses a unified `users` table with two roles: `admin` and `visitor`. Backend access is guarded by `authorizeAdmin`, which checks `role = admin` in the access token. The database also has a partial unique index that keeps the system to one admin:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS users_single_admin_idx ON users(role) WHERE role = 'admin';
```

The current user model has `deleted_at`, but it is not a good fit for account disabling because deletion and disabling are different product states. Login currently loads a user by username and creates tokens without checking a disabled state. Refresh token renewal already checks `sessions.is_blocked`, so disabling a user can reuse that mechanism for existing sessions.

## Data Model

Add explicit account-disable fields to `users`:

```sql
ALTER TABLE users
ADD COLUMN disabled_at timestamptz,
ADD COLUMN disabled_reason text NOT NULL DEFAULT '';
```

Semantics:

- `disabled_at IS NULL` means the visitor account is enabled.
- `disabled_at IS NOT NULL` means the visitor account is disabled.
- `disabled_reason` is optional user-facing admin context, not required for disabling.
- `deleted_at` remains reserved for future deletion semantics and is not used for disabling.

Indexes:

- Keep the existing `users_role_idx`.
- Add an index for visitor status and listing if needed:

```sql
CREATE INDEX IF NOT EXISTS users_role_disabled_created_idx
ON users(role, disabled_at, created_at DESC);
```

## Backend API

Add admin-only gRPC-Gateway endpoints under `/v1/users`. All endpoints call `authorizeAdmin`.

### List Users

`GET /v1/users`

Query parameters:

- `q`: optional search keyword.
- `status`: `all`, `enabled`, or `disabled`; defaults to `all`.
- `page`: minimum `1`; defaults to `1`.
- `limit`: allowed values `10`, `20`, `50`; defaults to `20`.

Behavior:

- Always filters `role = 'visitor'`.
- Searches username, full name, and email.
- Returns users plus total count.
- Never returns password hash.

### Update User

`PATCH /v1/users/{id}`

Editable fields:

- `full_name`
- `email`
- `is_email_verified`

Behavior:

- Only allows target users with `role = 'visitor'`.
- Returns `404` for missing users.
- Returns a permission or invalid-argument error if the target is not a visitor.
- Maps unique email conflicts to a readable error.

### Disable User

`POST /v1/users/{id}/disable`

Body:

- `reason`: optional string.

Behavior:

- Only allows target users with `role = 'visitor'`.
- Sets `disabled_at = now()` and stores `disabled_reason`.
- Blocks all sessions for the user with `is_blocked = true`.
- Is idempotent: disabling an already disabled user returns success with the current disabled state.
- Historical comments and other public content remain visible.

### Enable User

`POST /v1/users/{id}/enable`

Behavior:

- Only allows target users with `role = 'visitor'`.
- Clears `disabled_at` and `disabled_reason`.
- Is idempotent: enabling an already enabled user returns success.
- Does not automatically unblock old sessions. The restored user logs in again and receives fresh tokens.

## Authentication Changes

Public login must reject disabled users:

- After `GetUserByUsername`, check `disabled_at`.
- If disabled, return `401` or `403` with a readable message such as "account disabled".
- Do this before token/session creation.

Refresh behavior:

- On disable, block the user's sessions.
- Existing refresh token attempts fail through the existing `session.IsBlocked` check.
- Existing short-lived access tokens may remain valid until expiry. This is acceptable for the first version and avoids adding database reads to every authenticated request.

## Frontend Design

Add a "用户" item to the admin sidebar, after "分类" and before "AI 设置".

Route:

- `/backend/users`
- route name: `adminUsers`

Page layout:

- Header with title "用户管理" and a total count badge.
- Search input for username, email, and full name.
- Status segmented filter: all, enabled, disabled.
- Page size selector: `10 / 20 / 50`.
- Table with columns:
  - username
  - full name
  - email
  - email verification status
  - account status
  - created date
  - actions
- Pagination with previous/next, current page, total pages, and direct page jump.

Actions:

- Edit opens a compact modal.
- Disable opens a confirmation dialog with an optional reason field.
- Enable opens a confirmation dialog.
- After mutations, refresh the current page. If the current page becomes empty and page > 1, move back one page.

Visual rules:

- Use the existing admin surface/card/table styling.
- Keep the table dense and scannable.
- Use badges for enabled/disabled and email verified/unverified states.
- Avoid creating a separate detail page until future audit/session features exist.

## Error Handling

Backend:

- Missing user: `NotFound`.
- Target is not visitor: `InvalidArgument` or `PermissionDenied`.
- Duplicate email: user-readable conflict error.
- Invalid pagination: normalize page to `1`, clamp limit to allowed values.
- Disable/enable repeated action: return success.

Frontend:

- Use the existing admin HTTP error toast behavior.
- Show inline validation for empty full name or invalid email.
- Disable action buttons while requests are in flight.
- Keep destructive actions behind confirmation.

## Testing

Backend:

- SQL/sqlc tests for visitor listing, search, status filtering, pagination, disable, enable, and session blocking.
- gapi tests for admin authorization, visitor-only filtering, update fields, duplicate email handling, disable/enable behavior, and rejection of admin targets.
- login tests to ensure disabled users cannot log in.
- refresh tests to ensure blocked sessions cannot refresh.

Frontend:

- Contract tests that the admin route and sidebar include user management.
- API helper tests or source contract tests for list/update/disable/enable calls.
- UI contract tests for search, status filter, page size selector, page jump, edit, disable, and enable affordances.

Verification:

```bash
make sqlc
make proto
go test ./db/sqlc ./gapi ./api -count=1
cd web/frontend && bun run build
```

## Rollout Notes

This feature requires a database migration. Existing users should default to enabled because `disabled_at` is nullable. No data backfill is required.

The module deliberately preserves the single-admin model. If the product later needs multiple admins, that should be handled as a separate role-management design because it changes database constraints, permission semantics, and self-protection rules.
