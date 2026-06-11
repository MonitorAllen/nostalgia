# Platform Hardening Roadmap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the six identified optimization areas into a tracked roadmap and begin with CI coverage hardening.

**Architecture:** Treat each optimization as an independently shippable phase. The first branch upgrades GitHub Actions so backend, frontend, and deployment config regressions are caught before later performance, security, health-check, and database cleanup work lands.

**Tech Stack:** GitHub Actions, Go, PostgreSQL service containers, Bun, Vite, Docker Compose, Nginx, Vue.

---

## Roadmap Phases

1. **CI workflow coverage and env cleanup**
   - Run backend tests on every PR and push to `master`.
   - Run frontend `bun test`, `bun run type-check`, and `bun run build`.
   - Run `docker compose config --quiet` for production and dev compose files.
   - Stop decrypting `.env.dev.enc` in CI; use explicit CI environment variables instead.

2. **Frontend bundle splitting**
   - Split CKEditor, Prism/content rendering, and admin editor chunks with Vite `manualChunks`.
   - Keep route-level lazy loading for admin/editor surfaces.
   - Add build/config guard tests so large editor dependencies do not re-enter the public entry chunk.

3. **Frontend XSS and CSP hardening**
   - Centralize rich-text sanitization policy.
   - Add tests for dangerous HTML, URL protocols, and article/comment rendering.
   - Add a baseline Nginx Content-Security-Policy compatible with CKEditor output and uploaded resources.
   - Remove stray production `console.log` calls.

4. **Health checks and deployment observability**
   - Add API `/healthz` and `/readyz` endpoints.
   - Add Docker Compose healthchecks for `api`, `web`, `postgres`, and `redis`.
   - Add Nginx routing for health endpoints and document expected operational signals.

5. **Database legacy structure audit**
   - Audit migrations, generated sqlc code, queries, docs, and frontend assumptions for old admin/RBAC tables.
   - Remove unreachable legacy code only after confirming no migration/data path depends on it.
   - Document any manual data review needed before destructive migrations.

6. **Deployment and build polish**
   - Add Docker BuildKit registry cache when the registry supports it.
   - Add image labels and build metadata.
   - Consider `docker/build-push-action` after the plain shell workflow is stable.

## Phase 1 File Structure

- Modify `.github/workflows/test.yml`: split backend/frontend/compose jobs and remove `make decrypt_env env=dev`.
- Modify `web/frontend/src/deploy/nginxConfig.test.ts`: add CI workflow guard tests.
- Modify docs if verification commands or CI expectations change.

## Phase 1 Tasks

### Task 1: Add CI Workflow Guard Tests

**Files:**
- Modify: `web/frontend/src/deploy/nginxConfig.test.ts`

- [ ] **Step 1: Write failing tests**

Add tests that assert `.github/workflows/test.yml`:

```ts
expect(workflow).not.toContain('paths-ignore:')
expect(workflow).not.toContain('make decrypt_env env=dev')
expect(workflow).toContain('bun test')
expect(workflow).toContain('bun run type-check')
expect(workflow).toContain('bun run build')
expect(workflow).toContain('docker compose config --quiet')
expect(workflow).toContain('docker compose -f docker-compose.dev.yaml config --quiet')
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd web/frontend && bun test src/deploy/nginxConfig.test.ts
```

Expected: FAIL because current workflow ignores `web/**`, decrypts env, and lacks frontend/compose jobs.

### Task 2: Refactor Test Workflow

**Files:**
- Modify: `.github/workflows/test.yml`

- [ ] **Step 1: Update triggers**

Remove `paths-ignore` so frontend, compose, and backend changes all run tests.

- [ ] **Step 2: Replace env decryption**

Set explicit CI environment variables for backend tests:

```yaml
ENVIRONMENT: test
DB_DRIVER: postgres
DB_USER: root
DB_PASSWORD: secret
DB_SOURCE: postgresql://root:secret@localhost:5432/nostalgia?sslmode=disable
DB_URL: postgresql://root:secret@localhost:5432/nostalgia?sslmode=disable
MIGRATION_URL: file://db/migration
RESOURCE_PATH: ./resources
DOMAIN: http://localhost:8080
HTTP_SERVER_ADDRESS: 0.0.0.0:8080
GRPC_GATEWAY_ADDRESS: 0.0.0.0:9091
GRPC_SERVER_ADDRESS: 0.0.0.0:9090
TOKEN_SYMMETRIC_KEY: 12345678901234567890123456789012
SETUP_TOKEN: ci-setup-token
ACCESS_TOKEN_DURATION: 15m
REFRESH_TOKEN_DURATION: 24h
REDIS_ADDRESS: localhost:6379
EMAIL_SENDER_NAME: Nostalgia CI
EMAIL_SENDER_ADDRESS: noreply@example.com
EMAIL_SENDER_PASSWORD: ci-mail-password
UPLOAD_FILE_SIZE_LIMIT: 5242880
UPLOAD_FILE_ALLOWED_MIME: image/jpeg,image/png
HTTP_PROXY_ADDR: ""
```

- [ ] **Step 3: Add frontend job**

Use Bun setup, install frontend dependencies, then run:

```bash
cd web/frontend
bun install --frozen-lockfile
bun test
bun run type-check
bun run build
```

- [ ] **Step 4: Add compose config job**

Create a temporary `.env` from `.env.example` without real secrets and run:

```bash
docker compose config --quiet
docker compose -f docker-compose.dev.yaml config --quiet
```

### Task 3: Verify Phase 1

**Files:**
- Modified files above.

- [ ] **Step 1: Run targeted guard test**

Run:

```bash
cd web/frontend && bun test src/deploy/nginxConfig.test.ts
```

Expected: PASS.

- [ ] **Step 2: Run backend test suite**

Run:

```bash
make test
```

Expected: PASS.

- [ ] **Step 3: Run frontend checks**

Run:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: PASS; the existing Vite chunk-size warning may remain until Phase 2.

- [ ] **Step 4: Run compose checks**

Run:

```bash
docker compose config --quiet
docker compose -f docker-compose.dev.yaml config --quiet
```

Expected: PASS; local missing-variable warnings are acceptable if `.env` is absent.

- [ ] **Step 5: Commit phase 1**

Commit:

```bash
git add .github/workflows/test.yml web/frontend/src/deploy/nginxConfig.test.ts docs/superpowers/plans/2026-06-11-platform-hardening-roadmap.md
git commit -m "ci: cover frontend and compose checks"
```

## Phase 2 File Structure

- Modify `web/frontend/vite.config.ts`: add Rollup `manualChunks` rules for CKEditor, content rendering, and admin editor modules.
- Modify `web/frontend/src/router/index.ts`: lazy-load article detail routes so Prism/content rendering code is not pulled into the app shell route table.
- Add `web/frontend/src/deploy/viteBuildConfig.test.ts`: guard the bundle-splitting contract.

## Phase 2 Tasks

### Task 1: Add Bundle Splitting Guard Tests

- [x] Add a Vite config guard that requires manual chunks for CKEditor, Prism/content rendering, and admin editor code.
- [x] Add a router guard that prevents article detail from being statically imported into the public route table.
- [x] Verify the new guard fails before implementation with `cd web/frontend && bun test src/deploy/viteBuildConfig.test.ts`.

### Task 2: Split Frontend Bundles

- [x] Add `manualChunks` rules that isolate CKEditor dependencies in `ckeditor`.
- [x] Add `manualChunks` rules that isolate Prism, DOMPurify, and sanitization in `content-rendering`.
- [x] Add `manualChunks` rules that isolate admin article editor shell code in `admin-editor`.
- [x] Lazy-load `ArticleView.vue` from the router so article highlighting code loads on demand.

### Task 3: Verify Phase 2

- [x] Run `cd web/frontend && bun test`.
- [x] Run `cd web/frontend && bun run type-check`.
- [x] Run `cd web/frontend && bun run build`.
- [x] Confirm build output includes separate `ArticleView`, `content-rendering`, `admin-editor`, and `ckeditor` chunks.

## Phase 3 File Structure

- Add `web/frontend/src/util/sanitizeHtmlPolicy.ts`: central rich-text sanitization policy for URL protocols, target link rel values, and profile-aware DOMPurify config.
- Modify `web/frontend/src/util/sanitizeHtml.ts`: apply the centralized sanitization policy from the DOMPurify hook.
- Add `web/frontend/src/util/sanitizeHtmlPolicy.test.ts`: guard dangerous HTML/URL policy behavior.
- Add `web/frontend/src/deploy/frontendSourceHygiene.test.ts`: prevent stray production `console.log`/`console.debug`/`console.info` calls.
- Add `web/security-headers.conf`: shared Nginx security header and CSP include.
- Modify `web/nginx.conf`, `web/Dockerfile`, and `web/Dockerfile.dev`: include and ship the CSP/security header file.

## Phase 3 Tasks

### Task 1: Add XSS, CSP, and Console Guard Tests

- [x] Add sanitizer policy tests for allowed URL protocols and same-origin references.
- [x] Add sanitizer policy tests rejecting `javascript:`, obfuscated `javascript:`, `data:`, `vbscript:`, and `ftp:` URLs.
- [x] Add sanitizer policy tests for `rel="noopener noreferrer"` normalization and article/comment profiles.
- [x] Add Nginx deployment tests requiring a shipped baseline CSP include.
- [x] Add frontend source hygiene tests blocking stray production `console.log`/`console.debug`/`console.info` calls.
- [x] Verify the new tests fail before implementation.

### Task 2: Harden Frontend Rich Text and Nginx Headers

- [x] Centralize DOMPurify config and URL/rel policy, preserving CKEditor article styles while making comment rendering stricter.
- [x] Apply the policy from the existing DOMPurify `afterSanitizeAttributes` hook.
- [x] Add an Nginx Content-Security-Policy baseline compatible with app scripts, CKEditor styles, uploaded/self-hosted assets, and HTTPS images/media.
- [x] Include security headers in cacheable/static response locations so Nginx header inheritance cannot drop them.
- [x] Copy the security header include into production and development web images.
- [x] Remove stray production `console.log` calls.

### Task 3: Verify Phase 3

- [x] Run `cd web/frontend && bun test`.
- [x] Run `cd web/frontend && bun run type-check`.
- [x] Run `cd web/frontend && bun run build`.
- [x] Run Nginx syntax validation with mounted `web/nginx.conf`, `web/security-headers.conf`, and temporary local certificates.
- [x] Run `docker compose config --quiet && docker compose -f docker-compose.dev.yaml config --quiet`.

## Phase 4 File Structure

- Add `api/health.go`: Gin liveness and readiness endpoints.
- Add `api/health_test.go`: API behavior tests for `/healthz` and `/readyz`.
- Modify `db/sqlc/store.go`: expose `Ping(ctx)` on `Store` for database readiness.
- Modify `internal/cache/cache.go` and `internal/cache/redis.go`: expose Redis `Ping(ctx)` through the cache interface.
- Regenerate `db/mock/store.go` and `internal/cache/mock/redis.go`.
- Modify `web/nginx.conf`: route HTTPS health endpoints to API and keep local HTTP `/healthz` for web container checks.
- Modify `docker-compose.yaml` and `docker-compose.dev.yaml`: add healthchecks and healthy dependency conditions.
- Add `docs/deployment-healthchecks.md`: document operational signals.

## Phase 4 Tasks

### Task 1: Add Healthcheck Guard Tests

- [x] Add API tests for `/healthz`.
- [x] Add API tests for `/readyz` success, database failure, and Redis failure.
- [x] Add deployment tests for Nginx `/healthz` and `/readyz` routing.
- [x] Add deployment tests for Compose healthchecks on `postgres`, `redis`, `api`, and `web`.
- [x] Verify the new tests fail before implementation.

### Task 2: Implement Health and Readiness Signals

- [x] Add database `Ping(ctx)` to the store abstraction.
- [x] Add Redis `Ping(ctx)` to the cache abstraction.
- [x] Implement `GET /healthz` without external dependency checks.
- [x] Implement `GET /readyz` with PostgreSQL and Redis checks.
- [x] Regenerate Go mocks after interface changes.
- [x] Add Nginx health endpoint routing.
- [x] Add Docker Compose healthchecks and `condition: service_healthy` dependencies.
- [x] Document expected operational signals.

### Task 3: Verify Phase 4

- [x] Run targeted API health tests.
- [x] Run targeted Nginx/Compose guard tests.
- [x] Run `make test`.
- [x] Run `cd web/frontend && bun test`.
- [x] Run `cd web/frontend && bun run type-check`.
- [x] Run `cd web/frontend && bun run build`.
- [x] Run Nginx syntax validation with mounted config and temporary local certificates.
- [x] Run `docker compose config --quiet && docker compose -f docker-compose.dev.yaml config --quiet`.

## Phase 5 File Structure

- Add `web/frontend/src/deploy/databaseLegacyAudit.test.ts`: guard current schema docs, frontend admin types, and sqlc runtime models against removed admin/RBAC tables.
- Modify `doc/db.dbml` and `doc/schema.sql`: update current database documentation after `000009_remove_legacy_admin_rbac`.
- Modify `web/frontend/src/admin/types.ts`: remove legacy `role_id` from admin user types and model the unified `role` string.

## Phase 5 Tasks

### Task 1: Audit Legacy Admin/RBAC References

- [x] Scan migrations, sqlc output, db queries, frontend admin code, and docs for `admins`, `roles`, `role_permissions`, `sys_menus`, `role_id`, and related legacy RBAC assumptions.
- [x] Confirm runtime sqlc models and queries no longer expose legacy admin/RBAC tables.
- [x] Confirm historical migrations still mention old tables because they are part of the migration chain and should not be rewritten.
- [x] Identify stale current docs in `doc/db.dbml` and `doc/schema.sql`.
- [x] Identify stale frontend admin API type `role_id`.

### Task 2: Add Legacy Structure Guard Tests

- [x] Add tests proving current schema docs do not expose removed admin/RBAC tables.
- [x] Add tests proving frontend admin API types no longer model `role_id`.
- [x] Add tests proving sqlc runtime models do not include legacy admin/RBAC structs.
- [x] Verify the new tests fail before implementation.

### Task 3: Clean Current Schema Docs and Types

- [x] Update `doc/db.dbml` to current post-migration tables.
- [x] Update `doc/schema.sql` to current post-migration SQL structure.
- [x] Remove `role_id` from frontend admin types and use unified `role?: string`.

### Task 4: Verify Phase 5

- [x] Run targeted legacy audit tests.
- [x] Run `cd web/frontend && bun test`.
- [x] Run `cd web/frontend && bun run type-check`.
- [x] Run `make test`.
