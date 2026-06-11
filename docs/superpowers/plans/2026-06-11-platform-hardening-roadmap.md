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
