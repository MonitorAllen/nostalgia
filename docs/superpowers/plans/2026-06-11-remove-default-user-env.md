# Remove Default User Env Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove obsolete default-user and LetsEncrypt environment variables now that setup creates the first admin user.

**Architecture:** Setup remains the only first-admin bootstrap path. Admin article creation uses the authenticated JWT payload user ID as the article owner, and config loading no longer exposes `DEFAULT_USER_*` fields.

**Tech Stack:** Go, Gin/gRPC-Gateway, sqlc, GoMock, Viper, Bun config tests.

---

## File Structure

- Create `gapi/rpc_create_article_test.go`: regression test that `CreateArticle` uses the authenticated admin user ID as `owner`.
- Modify `gapi/rpc_create_article.go`: consume `authorizeAdmin` payload and remove `DefaultUserID` parsing.
- Modify `main.go`: remove `ensureDefaultUserExists` call and helper.
- Modify `util/config.go`: delete default-user config fields.
- Modify `util/config_test.go`: stop setting obsolete env vars and add a guard that removed fields are not present on `Config`.
- Modify `web/frontend/src/deploy/nginxConfig.test.ts`: add documentation/config guard against obsolete env vars.
- Modify `.env.example` and `README.md`: remove `DEFAULT_USER_*` and `LETSENCRYPT_EMAIL`.

## Tasks

### Task 1: Article Owner Uses Authenticated Admin

**Files:**
- Create: `gapi/rpc_create_article_test.go`
- Modify: `gapi/rpc_create_article.go`

- [ ] **Step 1: Write failing gRPC test**

Create a test that signs an admin token for `adminID`, leaves default-user config unset, calls `CreateArticle`, and expects `store.CreateArticle` to receive `Owner: adminID`.

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./gapi -run TestCreateArticleUsesAuthenticatedAdminAsOwner -count=1
```

Expected: FAIL because current implementation tries to parse `server.config.DefaultUserID` before creating the article.

- [ ] **Step 3: Implement owner change**

Capture `payload` from `authorizeAdmin(ctx)` and set `Owner: payload.UserID`.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./gapi -run TestCreateArticleUsesAuthenticatedAdminAsOwner -count=1
```

Expected: PASS.

### Task 2: Remove Default User Bootstrap Config

**Files:**
- Modify: `main.go`
- Modify: `util/config.go`
- Modify: `util/config_test.go`

- [ ] **Step 1: Write failing config guard**

Add a test that reflects over `util.Config` and asserts these mapstructure keys are absent:

```text
DEFAULT_USER_ID
DEFAULT_USERNAME
DEFAULT_USER_PASSWORD
DEFAULT_USER_FULLNAME
DEFAULT_USER_EMAIL
```

- [ ] **Step 2: Run config test to verify it fails**

Run:

```bash
go test ./util -run TestConfigDoesNotExposeDefaultUserBootstrapEnv -count=1
```

Expected: FAIL because `Config` still exposes the old fields.

- [ ] **Step 3: Remove bootstrap code and config fields**

Remove `ensureDefaultUserExists` invocation/helper from `main.go`, delete default-user fields from `util.Config`, and remove obsolete env entries from `util/config_test.go`.

- [ ] **Step 4: Run targeted Go tests**

Run:

```bash
go test ./util -run 'TestConfig|TestLoadConfig' -count=1
go test ./gapi -run TestCreateArticleUsesAuthenticatedAdminAsOwner -count=1
```

Expected: PASS.

### Task 3: Remove Obsolete Env From Docs And Examples

**Files:**
- Modify: `.env.example`
- Modify: `README.md`
- Modify: `web/frontend/src/deploy/nginxConfig.test.ts`

- [ ] **Step 1: Write failing repository guard**

Add a deployment/config test that scans `.env.example` and `README.md` and asserts none of these keys remain:

```text
DEFAULT_USER_ID
DEFAULT_USERNAME
DEFAULT_USER_PASSWORD
DEFAULT_USER_FULLNAME
DEFAULT_USER_EMAIL
LETSENCRYPT_EMAIL
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd web/frontend && bun test src/deploy/nginxConfig.test.ts
```

Expected: FAIL because docs/examples still list obsolete variables.

- [ ] **Step 3: Remove obsolete docs/example entries**

Delete those variables from `.env.example` and the README environment block.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
cd web/frontend && bun test src/deploy/nginxConfig.test.ts
```

Expected: PASS.

### Task 4: Verification And Commits

**Files:**
- All changed files above.

- [ ] **Step 1: Format Go files**

Run:

```bash
gofmt -w main.go util/config.go util/config_test.go gapi/rpc_create_article.go gapi/rpc_create_article_test.go
```

- [ ] **Step 2: Run backend verification**

Run:

```bash
make test
```

Expected: PASS.

- [ ] **Step 3: Run frontend/config verification**

Run:

```bash
cd web/frontend && bun test
```

Expected: PASS.

- [ ] **Step 4: Commit in logical slices**

Use separate commits for article owner behavior, bootstrap config removal, and docs/example cleanup.
