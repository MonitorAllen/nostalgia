# Auth Unification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Unify public and admin authentication on `users.role`, add one-time setup, upgrade JWT to `golang-jwt/jwt/v5 v5.3.1`, and remove legacy admin RBAC tables.

**Architecture:** Keep `/api` public auth as the source of login and refresh truth. Keep admin content APIs under `/v1` during this migration, but make gRPC-Gateway verify the same JWT user payload and enforce `role == admin`. Keep PASETO maker code as an optional backend, but remove admin-specific token/session flow from the default path.

**Tech Stack:** Go, Gin, gRPC-Gateway, sqlc, PostgreSQL migrations, JWT v5, Vue 3, Pinia, Bun.

---

## File Map

- `go.mod`, `go.sum`: replace `github.com/dgrijalva/jwt-go` with `github.com/golang-jwt/jwt/v5 v5.3.1`; bump vulnerable `grpc`, `x/net`, and `go-redis` to fixed versions if tests stay green.
- `token/payload.go`, `token/maker.go`, `token/jwt_maker.go`, `token/paseto_maker.go`, `token/*_test.go`: unified payload and JWT v5 implementation; keep PASETO compiling without default admin usage.
- `util/config.go`, `.env.example`: add `SETUP_TOKEN`.
- `db/query/user.sql`, `db/migration/000008_*`, `db/migration/000009_*`, `db/sqlc/*`, `db/mock/store.go`: role/setup queries, role constraint, legacy table cleanup, regenerated sqlc and mocks.
- `api/setup.go`, `api/setup_test.go`, `api/server.go`, `api/user.go`, `api/middleware.go`, `api/middleware_test.go`: setup API, public registration role safety, role middleware.
- `gapi/server.go`, `gapi/authorization.go`, `gapi/rpc_*admin*.go`, `gapi/rpc_init_sys_menu.go`, `gapi/*_test.go`: admin APIs verify unified JWT and stop using `admins`, `roles`, `role_permissions`, `sys_menus`.
- `proto/*admin*.proto`, `proto/rpc_init_sys_menu.proto`, `proto/service_nostalgia.proto`, `pb/*`: remove admin login/renew/info/menu RPCs if no frontend/backend code needs them; keep article/category/upload RPCs under `/v1`.
- `web/frontend/src/store/module/auth.ts`, `web/frontend/src/types/user.ts`, `web/frontend/src/admin/api/adminHttp.ts`, `web/frontend/src/admin/api/adminAuthApi.ts`, `web/frontend/src/admin/stores/adminAuth.ts`, `web/frontend/src/router/index.ts`, `web/frontend/src/views/admin/*.vue`, `web/frontend/src/views/setup/SetupView.vue`: unified frontend auth and setup page.

---

### Task 1: JWT v5 And Token Backend

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `token/payload.go`
- Modify: `token/maker.go`
- Modify: `token/jwt_maker.go`
- Modify: `token/paseto_maker.go`
- Modify: `token/jwt_maker_test.go`
- Test: `token/jwt_maker_test.go`

- [ ] **Step 1: Write/adjust JWT tests first**

Add or keep tests covering:

```go
func TestJWTMaker(t *testing.T)
func TestExpiredJWTToken(t *testing.T)
func TestInvalidJWTTokenAlgNone(t *testing.T)
func TestJWTMakerRejectsShortSecret(t *testing.T)
```

Expected behavior:
- valid token verifies and preserves `user_id`, `username`, `role`, `issued_at`, `expire_at`
- expired token returns `ErrExpiredToken`
- `alg=none` token returns `ErrInvalidToken`
- short secret rejects with `"invalid key size"`

- [ ] **Step 2: Run red/compat check**

Run:

```bash
go test ./token -run 'TestJWT|TestExpiredJWT' -count=1
```

Expected before implementation: either old tests pass on old dependency or fail after import changes; continue only after the test expectations are explicit.

- [ ] **Step 3: Upgrade JWT dependency**

Run:

```bash
go get github.com/golang-jwt/jwt/v5@v5.3.1
go mod tidy
```

Then remove `github.com/dgrijalva/jwt-go` from `go.mod`.

- [ ] **Step 4: Implement JWT v5 maker**

Use `github.com/golang-jwt/jwt/v5`.

Implementation shape:

```go
jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
signed, err := jwtToken.SignedString([]byte(maker.secretKey))
```

For verification:

```go
claims := &Payload{}
parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
    if token.Method != jwt.SigningMethodHS256 {
        return nil, ErrInvalidToken
    }
    return []byte(maker.secretKey), nil
}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
```

Map `jwt.ErrTokenExpired` to `ErrExpiredToken`; all other parse/signature/claim errors become `ErrInvalidToken`.

- [ ] **Step 5: Remove admin token methods from the default interface**

Update `token.Maker` to expose only:

```go
CreateToken(userID uuid.UUID, username string, role string, duration time.Duration) (string, *Payload, error)
VerifyToken(token string) (*Payload, error)
```

Keep PASETO compiling by implementing the same interface. Delete `AdminPayload`, `CreateAdminToken`, and `VerifyAdminToken` usage from the default path.

- [ ] **Step 6: Verify**

Run:

```bash
go test ./token -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum token
git commit -m "chore(auth): upgrade jwt backend"
```

---

### Task 2: Database, Setup API, And Role Queries

**Files:**
- Modify: `util/config.go`
- Modify: `.env.example`
- Modify: `db/query/user.sql`
- Create: `db/migration/000008_unify_user_roles.up.sql`
- Create: `db/migration/000008_unify_user_roles.down.sql`
- Create: `api/setup.go`
- Create: `api/setup_test.go`
- Modify: `api/server.go`
- Modify: `api/user.go`
- Regenerate: `db/sqlc/*`
- Regenerate: `db/mock/store.go`

- [ ] **Step 1: Write failing setup API tests**

Create `api/setup_test.go` with tests:

```go
func TestSetupStatusAPI(t *testing.T)
func TestCreateSetupAdminAPI(t *testing.T)
```

Cases:
- no admin user: `GET /api/setup/status` returns `initialized=false`, `setup_available=true`
- existing admin user: status returns `initialized=true`, `setup_available=false`
- wrong setup token: `POST /api/setup/admin` returns unauthorized/forbidden
- correct setup token and no admin: creates user with `role=admin`
- existing admin: `POST /api/setup/admin` returns conflict and does not call create

- [ ] **Step 2: Add setup config**

Add to `util.Config`:

```go
SetupToken string `mapstructure:"SETUP_TOKEN"`
```

Add `SETUP_TOKEN=` to `.env.example` without a real secret.

- [ ] **Step 3: Add user queries**

Extend `db/query/user.sql`:

```sql
-- name: CountAdminUsers :one
SELECT count(*) FROM users WHERE role = 'admin';

-- name: CreateUserWithRole :one
INSERT INTO users (
    id,
    username,
    hashed_password,
    full_name,
    email,
    is_email_verified,
    role
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;
```

Keep public `CreateUser` role-free so public registration cannot choose a role.

- [ ] **Step 4: Add role migration**

`000008_unify_user_roles.up.sql`:

```sql
UPDATE users SET role = 'visitor' WHERE role NOT IN ('admin', 'visitor');
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'visitor'));
CREATE INDEX IF NOT EXISTS users_role_idx ON users(role);
```

`000008_unify_user_roles.down.sql`:

```sql
DROP INDEX IF EXISTS users_role_idx;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
```

- [ ] **Step 5: Generate sqlc and mocks**

Run:

```bash
make sqlc
make mock
```

- [ ] **Step 6: Implement setup API**

Add handlers:

```go
func (server *Server) setupStatus(ctx *gin.Context)
func (server *Server) createSetupAdmin(ctx *gin.Context)
```

Rules:
- count admins first
- if count > 0, status initialized and create returns `409`
- compare request token with `server.config.SetupToken`
- if setup token is empty during creation, return server configuration error
- create admin using `CreateUserWithRole` with `Role: util.Admin` and `IsEmailVerified: true`

- [ ] **Step 7: Register setup routes**

In `api/server.go` public group:

```go
public.GET("/setup/status", server.setupStatus)
public.POST("/setup/admin", server.createSetupAdmin)
```

- [ ] **Step 8: Verify**

Run:

```bash
go test ./api -run 'TestSetup|TestCreateUserAPI' -count=1
go test ./db/sqlc -run 'TestCreateUser|TestUpdateUser' -count=1
```

- [ ] **Step 9: Commit**

```bash
git add util/config.go .env.example db/query/user.sql db/migration db/sqlc db/mock api
git commit -m "feat(auth): add guarded setup admin flow"
```

---

### Task 3: gRPC-Gateway Admin Authorization Cleanup

**Files:**
- Modify: `gapi/server.go`
- Modify: `gapi/authorization.go`
- Modify: admin-protected `gapi/rpc_*.go`
- Modify/delete: `gapi/rpc_login_admin.go`
- Modify/delete: `gapi/rpc_renew_access_token.go`
- Modify/delete: `gapi/rpc_admin_info.go`
- Modify/delete: `gapi/rpc_init_sys_menu.go`
- Modify/delete: related gapi tests
- Modify: `proto/service_nostalgia.proto`
- Regenerate: `pb/*`

- [ ] **Step 1: Write/update auth tests**

Update gapi tests so admin-protected RPCs use JWT user payloads:

```go
func newContextWithUserBearerToken(t *testing.T, tokenMaker token.Maker, userID uuid.UUID, username string, role string, duration time.Duration) context.Context
```

Cases:
- admin role passes
- visitor role returns unauthenticated/permission denied
- missing authorization fails
- expired token fails

- [ ] **Step 2: Switch gapi server to JWT maker**

In `gapi/server.go`, use:

```go
tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
```

- [ ] **Step 3: Replace authorizeAdmin**

Change return payload from `*token.AdminPayload` to `*token.Payload` and verify with `VerifyToken`.

Reject non-admin:

```go
if payload.Role != util.Admin {
    return nil, "", fmt.Errorf("admin role required")
}
```

- [ ] **Step 4: Remove legacy auth RPC usage**

The frontend will no longer call `/v1/admin/login`, `/v1/admin/renew_access`, `/v1/admin/info`, or `/v1/menu/init`.

Remove these RPCs from `proto/service_nostalgia.proto` and regenerate with:

```bash
make proto
```

Delete or adjust Go handlers/tests that only served those legacy endpoints.

- [ ] **Step 5: Remove admin/session/menu cache dependencies**

Remove `AdminSession`, `GetAdminSessionKey`, `GetAdmin`, `GetAdminById`, `ListInitSysMenus`, `RoleID`, and `AdminID` dependencies from gapi code.

- [ ] **Step 6: Verify**

Run:

```bash
go test ./gapi -count=1
go test ./pb -count=1
```

- [ ] **Step 7: Commit**

```bash
git add gapi proto pb
git commit -m "refactor(auth): use unified user jwt for admin gateway"
```

---

### Task 4: Frontend Unified Admin Auth And Setup Page

**Files:**
- Modify: `web/frontend/src/types/user.ts`
- Modify: `web/frontend/src/store/module/auth.ts`
- Modify: `web/frontend/src/util/http.ts`
- Modify: `web/frontend/src/admin/api/adminHttp.ts`
- Modify: `web/frontend/src/admin/api/adminAuthApi.ts`
- Delete or simplify: `web/frontend/src/admin/stores/adminAuth.ts`
- Modify: `web/frontend/src/router/index.ts`
- Modify: `web/frontend/src/views/admin/AdminLoginView.vue`
- Modify: `web/frontend/src/views/admin/AdminLayout.vue`
- Create: `web/frontend/src/service/setupService.ts`
- Create: `web/frontend/src/views/setup/SetupView.vue`

- [ ] **Step 1: Add frontend types**

Ensure user type includes:

```ts
role?: 'admin' | 'visitor'
```

- [ ] **Step 2: Extend unified auth store**

Use `useAuthStore` for admin and public auth. Add:

```ts
const isAdmin = computed(() => currentUser.value?.role === 'admin')
const ensureAuthenticated = async () => boolean
const ensureAdminAuthenticated = async () => boolean
```

Use existing public token storage keys. Remove admin-specific token keys from active code.

- [ ] **Step 3: Update admin HTTP client**

`adminHttp` should read token and refresh through `useAuthStore`, but keep `baseURL: '/v1'` for admin content APIs.

- [ ] **Step 4: Update admin login**

`AdminLoginView.vue` should call unified login. After login:

```ts
if (user.role !== 'admin') {
  authStore.logout()
  show permission feedback
  return
}
```

- [ ] **Step 5: Add setup service and page**

`setupService.ts`:

```ts
getSetupStatus()
createSetupAdmin(payload)
```

`SetupView.vue` fields:
- setup token
- username
- password
- full name
- email

On success, route to `/admin/login`.

- [ ] **Step 6: Update router guards**

Add route:

```ts
{ path: '/setup', name: 'setup', component: () => import('@/views/setup/SetupView.vue'), meta: { hideNavbar: true, hideFooter: true } }
```

Admin guard uses `authStore.ensureAdminAuthenticated()`.

- [ ] **Step 7: Verify**

Run:

```bash
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

- [ ] **Step 8: Commit**

```bash
git add web/frontend
git commit -m "feat(frontend): use unified auth for admin setup"
```

---

### Task 5: Remove Legacy Admin RBAC Schema And Generated Code

**Files:**
- Delete: `db/query/admin.sql`
- Delete: `db/query/sys_menu.sql`
- Create: `db/migration/000009_remove_legacy_admin_rbac.up.sql`
- Create: `db/migration/000009_remove_legacy_admin_rbac.down.sql`
- Regenerate/delete: `db/sqlc/admin.sql.go`, `db/sqlc/sys_menu.sql.go`, stale admin/menu models
- Regenerate: `db/mock/store.go`
- Modify: `gapi/converter.go`

- [ ] **Step 1: Write cleanup migration**

Up:

```sql
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS sys_menus CASCADE;
DROP TABLE IF EXISTS admins CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
```

Down should recreate the old tables with the same structure from `000003_add_backend_module.up.sql` so rollback can run.

- [ ] **Step 2: Remove old queries and regenerate**

Run:

```bash
rm db/query/admin.sql db/query/sys_menu.sql
make sqlc
rm -f db/sqlc/admin.sql.go db/sqlc/sys_menu.sql.go
make mock
```

- [ ] **Step 3: Remove stale converter/admin references**

Delete `convertAdmin`, `convertInitSysMenu`, and menu tree helpers if no caller remains.

- [ ] **Step 4: Verify**

Run:

```bash
go test ./db/sqlc ./gapi ./api -count=1
```

- [ ] **Step 5: Commit**

```bash
git add db gapi
git commit -m "chore(auth): remove legacy admin rbac"
```

---

### Task 6: Dependency Security Verification

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Upgrade fixed dependency versions**

Run:

```bash
go get google.golang.org/grpc@v1.79.3
go get golang.org/x/net@v0.38.0
go get github.com/redis/go-redis/v9@v9.6.3
go mod tidy
```

- [ ] **Step 2: Verify vulnerability scan**

Run:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Expected: no reachable vulnerabilities. If new Go toolchain requirements block a latest scanner run, document the exact error and run the newest compatible scanner.

- [ ] **Step 3: Verify tests**

Run:

```bash
make test
```

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore(deps): address reachable go vulnerabilities"
```

---

## Final Verification

- [ ] Run `make sqlc` and confirm no generated diff remains.
- [ ] Run `make proto` and confirm generated pb files are committed.
- [ ] Run `make test`.
- [ ] Run `cd web/frontend && bun run type-check`.
- [ ] Run `cd web/frontend && bun run build`.
- [ ] Run `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`.
- [ ] Review `git diff master...HEAD --stat`.
- [ ] Push branch and open/update PR after verification.
