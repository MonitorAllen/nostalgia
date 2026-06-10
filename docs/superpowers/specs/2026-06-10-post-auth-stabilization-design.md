# Post-Auth Stabilization Design

## Context

Nostalgia now uses one user identity model for public and `/admin` traffic. Admin access is backed by `users.role = 'admin'`, public visitors use `users.role = 'visitor'`, and the first admin is created through the guarded setup flow.

That migration removed a large amount of legacy admin RBAC code. The next step is not a new feature; it is a stabilization pass so deployment, local browser state, and user-facing errors match the new model.

## Goals

- Document the new first-run setup flow in the main README.
- Make `.env.example` clearer about `SETUP_TOKEN` and local ports.
- Remove legacy auth localStorage keys as soon as the frontend auth store initializes.
- Keep `/admin/login` errors readable for owner use without exposing sensitive details.
- Avoid changing backend auth behavior unless a small test-backed gap is found.
- Keep this branch small enough to merge before starting CKEditor/editor parity work.

## Non-Goals

- No CKEditor redesign or article content style work.
- No token model changes.
- No new database migrations.
- No new admin management UI.
- No branch protection or GitHub settings work.

## Frontend Storage Stabilization

The auth store already writes unified user token keys:

```text
nostalgia_user_token
nostalgia_user_token_expires_at
nostalgia_user_refresh_token
nostalgia_user_refresh_token_expires_at
nostalgia_user_info
```

It also knows about legacy public/admin keys, but cleanup currently happens only when new tokens are written or the store is cleared. A returning browser can keep obsolete admin token keys until the user takes an auth action.

The store should remove known legacy keys during initialization. This cleanup must not remove the new unified keys, current user info, or unrelated localStorage values such as theme mode.

## Admin Login Error UX

The `/admin/login` page should continue using unified `/api/users/login`. If the returned user is not `admin`, it should clear tokens and display a concise permission message.

Network and backend errors should be shown as useful owner-facing messages:

- non-admin account: `当前账号没有后台权限`
- backend error text: use `error` or `message` from response data
- unknown login failure: `管理员账号或密码不正确`

No setup token, password, refresh token, or raw stack detail should be shown in the UI.

## Documentation

The README should describe:

- `SETUP_TOKEN` as a deployment-time bootstrap secret, not a reusable admin password.
- First-run sequence: configure env, migrate DB, start services, open `/setup`, create owner admin, then use `/admin/login`.
- Auth model summary: registered users are `visitor`; only the setup-created owner is `admin`.
- Frontend commands use Bun.

The `.env.example` file should keep placeholder values only. It should include a comment-free, copyable `SETUP_TOKEN=` key and use the current local PostgreSQL port from the Makefile if the project has intentionally moved away from default `5432`.

## Verification

Minimum verification for this branch:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
make test
```

If `bun test` is introduced only for the storage helper, it should be fast and not require a browser.
