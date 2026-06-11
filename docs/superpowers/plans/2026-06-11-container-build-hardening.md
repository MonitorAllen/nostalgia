# Container Build Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep production secrets out of Docker build contexts and images while tightening API/Web image build behavior.

**Architecture:** CI builds immutable images without decrypting production `.env`; runtime deployment remains responsible for injecting environment variables and mounted certificates. Dockerfiles use BuildKit cache mounts for dependency/build caches and expose the actual service ports.

**Tech Stack:** Docker BuildKit, GitHub Actions, Go 1.24 Alpine, Bun, Nginx, Bun tests.

---

## File Structure

- Modify `web/frontend/src/deploy/nginxConfig.test.ts`: add build/deploy guard tests beside the existing deployment config tests.
- Create `util/config_test.go`: cover env-only runtime configuration and `.env` override behavior.
- Modify `util/config.go`: allow runtime-only environment configuration when `.env` is absent.
- Modify `Dockerfile`: remove `.env` copy, add BuildKit syntax/cache mounts, expose `8080 9091`.
- Modify `.dockerignore`: ignore plaintext env files and local runtime artifacts so secrets cannot enter the API build context.
- Modify `docker-compose.yaml` and `docker-compose.dev.yaml`: inject API configuration at runtime via `env_file`.
- Modify `.github/workflows/deploy.yml`: remove `make decrypt_env env=prod` from image builds and enable BuildKit when building images.
- Modify `web/Dockerfile`: use Bun install cache and remove production self-signed cert generation.
- Modify `web/Dockerfile.dev`: use Bun install cache while keeping local self-signed cert generation.
- Delete `web/frontend/Dockerfile`: remove stale legacy frontend image path.
- Modify `README.md` and agent context docs if needed: document runtime secret injection and Cloudflare Origin CA mount expectation.

## Tasks

### Task 1: Add Build Guard Tests

**Files:**
- Modify: `web/frontend/src/deploy/nginxConfig.test.ts`

- [ ] **Step 1: Write failing tests**

Add tests that assert:

```ts
expect(readRepoFile('Dockerfile')).not.toMatch(/COPY\s+--from=builder\s+\/app\/\.env\b/)
expect(readRepoFile('.github/workflows/deploy.yml')).not.toContain('make decrypt_env env=prod')
expect(readRepoFile('.dockerignore')).toMatch(/^\.env$/m)
expect(readRepoFile('Dockerfile')).toContain('EXPOSE 8080 9091')
expect(readRepoFile('web/Dockerfile')).not.toContain('openssl req -x509')
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd web/frontend && bun test src/deploy/nginxConfig.test.ts
```

Expected: FAIL because current production build still copies `.env`, decrypts prod env, misses `.env` ignore, only exposes `8080`, and generates a fallback cert in production.

- [ ] **Step 3: Keep tests focused**

Only assert repository configuration content; do not read real `.env` values.

### Task 2: Harden API Image Build

**Files:**
- Create: `util/config_test.go`
- Modify: `util/config.go`
- Modify: `Dockerfile`
- Modify: `.dockerignore`
- Modify: `docker-compose.yaml`
- Modify: `docker-compose.dev.yaml`
- Modify: `.github/workflows/deploy.yml`

- [ ] **Step 1: Write runtime config tests**

Add Go tests proving `LoadConfig` works without a `.env` file when environment variables are provided, and that real environment variables override values from an existing `.env` file.

- [ ] **Step 2: Run config tests to verify they fail**

Run:

```bash
go test ./util -run 'TestLoadConfig' -count=1
```

Expected: FAIL because current `LoadConfig` returns an error when `.env` is missing.

- [ ] **Step 3: Update `LoadConfig`**

Use an isolated Viper instance, bind all `mapstructure` keys to environment variables, and ignore only missing `.env` file errors before unmarshalling.

- [ ] **Step 4: Run config tests to verify they pass**

Run:

```bash
go test ./util -run 'TestLoadConfig' -count=1
```

Expected: PASS.

- [ ] **Step 5: Update API Dockerfile**

Use Dockerfile syntax `# syntax=docker/dockerfile:1.7`, cache Go modules and build cache, remove the `.env` copy, and expose both HTTP and gRPC-Gateway ports:

```dockerfile
# syntax=docker/dockerfile:1.7
RUN --mount=type=cache,target=/go/pkg/mod \
    go env -w GO111MODULE=on && \
    go mod download
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main main.go
EXPOSE 8080 9091
```

- [ ] **Step 6: Update `.dockerignore`**

Ignore plaintext env files and local runtime folders:

```dockerignore
.env
.env.*
!*.env.example
certs/
postgres_data/
```

- [ ] **Step 7: Update Compose runtime env injection**

Add API `env_file` entries to both production and development Compose files:

```yaml
env_file:
  - path: .env
```

- [ ] **Step 8: Update deploy workflow**

Remove the `Decrypt ENV` step from `build-api`, enable BuildKit for `docker build`, and normalize the `latest` tag shell indentation.

### Task 3: Harden Web Image Build

**Files:**
- Modify: `web/Dockerfile`
- Modify: `web/Dockerfile.dev`
- Delete: `web/frontend/Dockerfile`

- [ ] **Step 1: Add Bun cache mounts**

Add Dockerfile syntax `# syntax=docker/dockerfile:1.7` and cache Bun installs:

```dockerfile
RUN --mount=type=cache,target=/root/.bun/install/cache bun install --frozen-lockfile
```

- [ ] **Step 2: Remove production fallback certificate generation**

In `web/Dockerfile`, keep `tzdata`, create `/etc/nginx/certs`, but do not install `openssl` or generate a self-signed cert. Production must mount Cloudflare Origin CA via Compose.

- [ ] **Step 3: Keep dev fallback certificate generation**

In `web/Dockerfile.dev`, keep `openssl` and self-signed cert generation so local HTTPS still works.

- [ ] **Step 4: Delete unused legacy frontend Dockerfile**

Remove `web/frontend/Dockerfile` after confirming no repo references require it.

### Task 4: Documentation And Verification

**Files:**
- Modify: `README.md`
- Modify: `.agents/skills/nostalgia-project/SKILL.md` if command guidance changes
- Modify: `AGENTS.md` only if deployment stack notes need adjustment

- [ ] **Step 1: Update docs**

State that production secrets are injected at runtime and Docker builds must not decrypt or copy `.env`.

- [ ] **Step 2: Run targeted tests**

Run:

```bash
cd web/frontend && bun test src/deploy/nginxConfig.test.ts
go test ./util -run 'TestLoadConfig' -count=1
```

Expected: PASS.

- [ ] **Step 3: Run frontend verification**

Run:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: PASS; existing Vite chunk-size warning may remain.

- [ ] **Step 4: Run compose/config checks**

Run:

```bash
docker compose config --quiet
docker compose -f docker-compose.dev.yaml config --quiet
```

Expected: PASS; local warnings about unset `.env` values are acceptable.

- [ ] **Step 5: Build images if local Docker is available**

Run:

```bash
DOCKER_BUILDKIT=1 docker build -t nostalgia-api:build-hardening .
DOCKER_BUILDKIT=1 docker build -t nostalgia-web:build-hardening web
```

Expected: PASS. If network or Docker daemon availability blocks this, report the exact blocker.
