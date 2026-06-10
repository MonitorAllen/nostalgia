# Cloudflare Nginx Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Caddy edge container with Nginx running in the `web` container, and make `/backend` the canonical admin frontend path.

**Architecture:** Keep API services internal to the Docker network. Let Nginx terminate Cloudflare Origin CA TLS, serve the Vue SPA and uploaded resources, and proxy `/api/*` and `/v1/*` to the Go service. Centralize frontend admin URLs in a small tested module so route and redirect changes are deliberate.

**Tech Stack:** Docker Compose, Nginx, Cloudflare Origin CA, Vue Router, TypeScript, Bun tests, Vite.

---

## File Map

- Create: `web/frontend/src/admin/adminRoutes.ts`
  - Canonical admin path constants and helpers for login redirects.
- Create: `web/frontend/src/admin/adminRoutes.test.ts`
  - Tests `/backend` route constants and login redirect behavior.
- Modify: `web/frontend/src/router/index.ts`
  - Changes admin route base from `/admin` to `/backend`.
- Modify: `web/frontend/src/admin/api/adminHttp.ts`
  - Uses shared admin login path helper.
- Modify: `web/frontend/src/admin/stores/adminAuth.ts`
  - Uses shared admin login path for logout.
- Modify: `web/frontend/src/views/admin/AdminLoginView.vue`
  - Uses shared admin articles path fallback.
- Modify: `web/frontend/src/views/setup/SetupView.vue`
  - Continues to use named admin route after route base changes.
- Create: `web/frontend/src/deploy/nginxConfig.test.ts`
  - Tests deployment text config for Caddy removal and Nginx ingress routing.
- Modify: `web/nginx.conf`
  - Adds Nginx ingress behavior, TLS server, API proxies, and removes `/backend -> /admin`.
- Modify: `docker-compose.yaml`
  - Removes Caddy, exposes `web` on `80/443`, mounts Cloudflare Origin certs.
- Modify: `docker-compose.dev.yaml`
  - Removes Caddy, exposes `web` on `80/443`, mounts Cloudflare Origin certs.
- Delete: `Caddyfile`
  - No longer used.
- Delete: `Caddyfile.dev`
  - No longer used.
- Modify: `README.md`
  - Updates deployment topology, `/backend` links, and Cloudflare Origin CA notes.

---

### Task 1: Admin Route Constants

**Files:**
- Create: `web/frontend/src/admin/adminRoutes.test.ts`
- Create: `web/frontend/src/admin/adminRoutes.ts`

- [ ] **Step 1: Write failing tests**

Create `web/frontend/src/admin/adminRoutes.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import {
  ADMIN_ARTICLES_PATH,
  ADMIN_BASE_PATH,
  ADMIN_LOGIN_PATH,
  buildAdminLoginRedirect,
} from './adminRoutes'

describe('adminRoutes', () => {
  test('uses backend as the canonical admin base path', () => {
    expect(ADMIN_BASE_PATH).toBe('/backend')
    expect(ADMIN_LOGIN_PATH).toBe('/backend/login')
    expect(ADMIN_ARTICLES_PATH).toBe('/backend/articles')
  })

  test('does not add a redirect query when already on the login page', () => {
    expect(buildAdminLoginRedirect('/backend/login')).toBe('/backend/login')
    expect(buildAdminLoginRedirect('/backend/login?redirect=%2Fbackend%2Farticles')).toBe(
      '/backend/login',
    )
  })

  test('preserves protected destination as an encoded redirect query', () => {
    expect(buildAdminLoginRedirect('/backend/articles?page=2#top')).toBe(
      '/backend/login?redirect=%2Fbackend%2Farticles%3Fpage%3D2%23top',
    )
  })
})
```

- [ ] **Step 2: Run red check**

Run:

```bash
cd web/frontend && bun test src/admin/adminRoutes.test.ts
```

Expected: FAIL because `adminRoutes.ts` does not exist.

- [ ] **Step 3: Implement route constants**

Create `web/frontend/src/admin/adminRoutes.ts`:

```ts
export const ADMIN_BASE_PATH = '/backend'
export const ADMIN_LOGIN_PATH = `${ADMIN_BASE_PATH}/login`
export const ADMIN_ARTICLES_PATH = `${ADMIN_BASE_PATH}/articles`

export const buildAdminLoginRedirect = (current: string) => {
  if (current.startsWith(ADMIN_LOGIN_PATH)) return ADMIN_LOGIN_PATH

  return `${ADMIN_LOGIN_PATH}?redirect=${encodeURIComponent(current)}`
}
```

- [ ] **Step 4: Run green check**

Run:

```bash
cd web/frontend && bun test src/admin/adminRoutes.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/admin/adminRoutes.ts web/frontend/src/admin/adminRoutes.test.ts
git commit -m "test(frontend): cover backend admin route constants"
```

### Task 2: Move Frontend Admin Routes to Backend

**Files:**
- Modify: `web/frontend/src/router/index.ts`
- Modify: `web/frontend/src/admin/api/adminHttp.ts`
- Modify: `web/frontend/src/admin/stores/adminAuth.ts`
- Modify: `web/frontend/src/views/admin/AdminLoginView.vue`
- Modify: `web/frontend/src/views/setup/SetupView.vue`

- [ ] **Step 1: Change Vue route base**

In `web/frontend/src/router/index.ts`, import constants:

```ts
import { ADMIN_BASE_PATH, ADMIN_LOGIN_PATH } from '@/admin/adminRoutes'
```

Change:

```ts
path: '/admin/login'
path: '/admin'
```

to:

```ts
path: ADMIN_LOGIN_PATH
path: ADMIN_BASE_PATH
```

- [ ] **Step 2: Update admin HTTP login redirect**

In `web/frontend/src/admin/api/adminHttp.ts`, import:

```ts
import { buildAdminLoginRedirect } from '@/admin/adminRoutes'
```

Replace `redirectToLogin()` body with:

```ts
private redirectToLogin() {
  const current = `${window.location.pathname}${window.location.search}${window.location.hash}`
  window.location.href = buildAdminLoginRedirect(current)
}
```

- [ ] **Step 3: Update admin logout path**

In `web/frontend/src/admin/stores/adminAuth.ts`, import:

```ts
import { ADMIN_LOGIN_PATH } from '@/admin/adminRoutes'
```

Change:

```ts
window.location.href = '/admin/login'
```

to:

```ts
window.location.href = ADMIN_LOGIN_PATH
```

- [ ] **Step 4: Update login fallback**

In `web/frontend/src/views/admin/AdminLoginView.vue`, import:

```ts
import { ADMIN_ARTICLES_PATH } from '@/admin/adminRoutes'
```

Replace both `'/admin/articles'` fallbacks with `ADMIN_ARTICLES_PATH`.

- [ ] **Step 5: Verify frontend route migration**

Run:

```bash
cd web/frontend && bun test src/admin/adminRoutes.test.ts
cd web/frontend && bun run type-check
```

Expected: all commands exit 0.

- [ ] **Step 6: Commit**

```bash
git add web/frontend/src/router/index.ts web/frontend/src/admin/api/adminHttp.ts web/frontend/src/admin/stores/adminAuth.ts web/frontend/src/views/admin/AdminLoginView.vue web/frontend/src/views/setup/SetupView.vue
git commit -m "refactor(frontend): use backend admin route base"
```

### Task 3: Nginx and Compose Ingress

**Files:**
- Create: `web/frontend/src/deploy/nginxConfig.test.ts`
- Modify: `web/nginx.conf`
- Modify: `docker-compose.yaml`
- Modify: `docker-compose.dev.yaml`
- Delete: `Caddyfile`
- Delete: `Caddyfile.dev`

- [ ] **Step 1: Write failing deployment config tests**

Create `web/frontend/src/deploy/nginxConfig.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const repoRoot = resolve(import.meta.dir, '../../../..')
const readRepoFile = (path: string) => readFileSync(resolve(repoRoot, path), 'utf8')

describe('nginx deployment ingress config', () => {
  test('proxies public and admin APIs from nginx', () => {
    const nginx = readRepoFile('web/nginx.conf')

    expect(nginx).toContain('proxy_pass http://api:8080')
    expect(nginx).toContain('proxy_pass http://api:9091')
    expect(nginx).not.toContain('return 308 /admin')
  })

  test('uses cloudflare origin certificate paths', () => {
    const nginx = readRepoFile('web/nginx.conf')

    expect(nginx).toContain('/etc/nginx/certs/cloudflare-origin.pem')
    expect(nginx).toContain('/etc/nginx/certs/cloudflare-origin.key')
  })

  test('production compose exposes web as the only public ingress', () => {
    const compose = readRepoFile('docker-compose.yaml')

    expect(compose).not.toContain('caddy:')
    expect(compose).toContain('"80:80"')
    expect(compose).toContain('"443:443"')
    expect(compose).toContain('./certs:/etc/nginx/certs:ro')
  })
})
```

- [ ] **Step 2: Run red check**

Run:

```bash
cd web/frontend && bun test src/deploy/nginxConfig.test.ts
```

Expected: FAIL because Nginx still lacks API proxies/TLS and Compose still contains `caddy`.

- [ ] **Step 3: Update Nginx ingress config**

Update `web/nginx.conf` so:

```nginx
server {
    listen 80 default_server;
    server_name _;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl default_server;
    server_name _;

    ssl_certificate /etc/nginx/certs/cloudflare-origin.pem;
    ssl_certificate_key /etc/nginx/certs/cloudflare-origin.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;

    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $http_cf_connecting_ip;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /v1/ {
        proxy_pass http://api:9091;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $http_cf_connecting_ip;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        root /usr/share/nginx/html;
        index index.html;
        try_files $uri $uri/ /index.html;
        add_header Cache-Control "no-store, no-cache, must-revalidate";
    }
}
```

Keep existing gzip, asset cache, `/resources/`, and `50x.html` behavior.

- [ ] **Step 4: Update Compose files**

In both Compose files:

- Remove the entire `caddy` service.
- Remove `caddy_data` and `caddy_config` volumes.
- Add to `web.ports`:

```yaml
      - "80:80"
      - "443:443"
```

- Add to `web.volumes`:

```yaml
      - ./certs:/etc/nginx/certs:ro
```

- [ ] **Step 5: Delete Caddy files**

Delete:

```bash
git rm Caddyfile Caddyfile.dev
```

- [ ] **Step 6: Run config tests**

Run:

```bash
cd web/frontend && bun test src/deploy/nginxConfig.test.ts
docker compose config
docker compose -f docker-compose.dev.yaml config
```

Expected: all commands exit 0.

- [ ] **Step 7: Commit**

```bash
git add web/frontend/src/deploy/nginxConfig.test.ts web/nginx.conf docker-compose.yaml docker-compose.dev.yaml Caddyfile Caddyfile.dev
git commit -m "chore(deploy): move edge routing to nginx"
```

### Task 4: Documentation and Final Verification

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README deployment docs**

Change README references:

- Project overview says `/backend` instead of `/admin`.
- Frontend directory says public blog and `/backend`.
- Quick deploy service list removes Caddy wording.
- Local Docker backend link becomes `http://localhost/backend/`.
- Setup completion login link becomes `http://localhost/backend/login`.
- Add a Cloudflare note:

```md
Production HTTPS assumes Cloudflare proxy with SSL/TLS mode set to Full (strict). Put the Cloudflare Origin CA certificate and key at `./certs/cloudflare-origin.pem` and `./certs/cloudflare-origin.key`; do not commit these files.
```

- [ ] **Step 2: Run full verification**

Run:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
docker compose config
docker compose -f docker-compose.dev.yaml config
```

Expected: all commands exit 0.

- [ ] **Step 3: Run optional Nginx syntax check**

If certificate files exist at `./certs/cloudflare-origin.pem` and `./certs/cloudflare-origin.key`, run:

```bash
docker run --rm -v "$PWD/web/nginx.conf:/etc/nginx/nginx.conf:ro" -v "$PWD/web/frontend/dist:/usr/share/nginx/html:ro" -v "$PWD/resources:/usr/share/nginx/resources:ro" -v "$PWD/certs:/etc/nginx/certs:ro" nginx:alpine nginx -t
```

Expected: PASS. If cert files do not exist locally, report that this verification is pending local Cloudflare Origin CA files.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs(deploy): document cloudflare nginx deployment"
```

- [ ] **Step 5: Push and open PR**

Run:

```bash
git status --short --branch
git push -u origin chore/cloudflare-nginx-deploy
gh pr create --base master --head chore/cloudflare-nginx-deploy --title "chore(deploy): simplify cloudflare nginx ingress" --body-file /tmp/nostalgia-cloudflare-nginx-pr.md
```
