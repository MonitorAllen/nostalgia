# Cache Layer Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Harden Redis-backed cache behavior, add bounded versioned article pagination cache, and split cache/queue Redis DB indexes.

**Architecture:** Keep one Redis service and select separate DB indexes for cache and Asynq queue clients. Add typed cache helpers around the existing low-level Redis primitive interface so handlers stop owning key and TTL policy. Use versioned article list keys plus singleflight to cache public list pages without scanning Redis on article writes.

**Tech Stack:** Go, Gin, gRPC/Gateway, sqlc, Redis via `github.com/redis/go-redis/v9`, Asynq, `golang.org/x/sync/singleflight`, GoMock tests.

---

## File Structure

- Modify `util/config.go`: add `REDIS_CACHE_DB` and `REDIS_QUEUE_DB` config fields with safe defaults.
- Modify `util/config_test.go`: cover Redis DB config defaults and env overrides.
- Modify `main.go`: pass queue DB to Asynq and cache DB to Redis cache.
- Modify `internal/cache/cache.go`: add typed-safe cache primitives required by versioned pagination, including `Incr` and corrected TTL inspection.
- Modify `internal/cache/redis.go`: select DB, implement `Incr`, fix TTL semantics.
- Modify `internal/cache/mock/redis.go`: regenerate cache mock after interface changes.
- Modify `internal/cache/key/article.go`: add distinct article ID/slug/list/idempotency keys.
- Add `internal/cache/ttl.go` and `internal/cache/ttl_test.go`: central TTL constants and jitter helper tests.
- Add `internal/cache/article.go` and `internal/cache/article_test.go`: article detail and list typed cache helpers.
- Add `internal/cache/category.go`: category typed cache helper.
- Add `internal/cache/contribution.go`: GitHub contribution typed cache helper.
- Add `internal/cache/idempotency.go`: like/view idempotency helper.
- Modify `api/article.go`: use `ArticleCache` and `IdempotencyCache`; add public list caching.
- Modify `api/article_test.go`: cover list page cache hit/miss/bounds.
- Modify `api/category.go` and `gapi/rpc_list_categories.go`: use `CategoryCache`.
- Modify `api/user.go`: cache raw GitHub contributions and make response slicing bounded.
- Modify `gapi/rpc_update_article.go` and `gapi/rpc_delete_article.go`: invalidate article detail keys and bump list versions.
- Modify `gapi/*category*.go`: use typed category invalidation if needed.
- Modify `worker/task_delay_delete_cache.go`: log delete failures per key.
- Modify `.env.example`, `docker-compose.yaml`, `docker-compose.dev.yaml`, `README.md`: document Redis DB split.

## Task 1: Redis DB Split and Primitive Semantics

- [x] **Step 1: Write failing config and cache primitive tests**

Add tests that expect `util.Config` to expose `RedisCacheDB` and `RedisQueueDB`, and tests that expect cache TTL inspection to distinguish missing keys from permanent keys.

Run:

```bash
go test ./util ./internal/cache
```

Expected: fail because the new config fields and cache primitive behavior are not implemented.

- [x] **Step 2: Implement config fields and Redis DB selection**

Add `RedisCacheDB int` and `RedisQueueDB int` to `util.Config`. Set defaults in `LoadConfig` before unmarshalling:

```go
configReader.SetDefault("REDIS_CACHE_DB", 0)
configReader.SetDefault("REDIS_QUEUE_DB", 1)
```

Update `cache.NewRedisCache` to use `redis.Options{Addr: config.RedisAddress, DB: config.RedisCacheDB}`. Update Asynq Redis options in `main.go` to set `DB: config.RedisQueueDB`.

- [x] **Step 3: Implement cache primitive updates**

Add `Incr(ctx, key)` and replace ambiguous expiration inspection with a method that can represent missing, expiring, and permanent keys. Regenerate `internal/cache/mock/redis.go` with:

```bash
make mock
```

- [x] **Step 4: Verify and commit**

Run:

```bash
gofmt -w util/config.go util/config_test.go internal/cache/cache.go internal/cache/redis.go main.go
go test ./util ./internal/cache
git add util/config.go util/config_test.go internal/cache/cache.go internal/cache/redis.go internal/cache/mock/redis.go main.go
git commit -m "feat: split redis cache and queue databases"
```

## Task 2: Key Naming and TTL Policy

- [x] **Step 1: Write failing key and TTL tests**

Add tests proving:

- Article ID key is `cache:article:id:{uuid}`.
- Article slug key is `cache:article:slug:{slug}`.
- Article list version keys use `all` or `category:{id}`.
- Article list page key includes version, category bucket, page, and limit.
- TTL jitter returns a value in the expected range.

Run:

```bash
go test ./internal/cache/...
```

Expected: fail because key formats and TTL helpers are not implemented.

- [x] **Step 2: Implement key constants and TTL helpers**

Update `internal/cache/key/article.go` with distinct key patterns and add list key helpers:

```go
func GetArticleListVersionKey(categoryID int64) string
func GetArticleListKey(version int64, categoryID int64, page int32, limit int32) string
```

Add `internal/cache/ttl.go` with constants for article detail, article list, empty list, category list, contributions, likes, and views.

- [x] **Step 3: Verify and commit**

Run:

```bash
gofmt -w internal/cache/key/article.go internal/cache/ttl.go internal/cache/*_test.go
go test ./internal/cache/...
git add internal/cache/key/article.go internal/cache/ttl.go internal/cache/*_test.go
git commit -m "feat: define cache keys and ttl policy"
```

## Task 3: Typed Cache Helpers and Public Article List Cache

- [x] **Step 1: Write failing typed cache tests**

Add tests for `ArticleCache` that prove:

- `GetList` returns false for pages greater than 5.
- `SetList` skips pages greater than 5.
- `SetList` uses short TTL for empty article lists.
- `BumpListVersion` increments the all bucket and each category bucket.

Run:

```bash
go test ./internal/cache
```

Expected: fail because typed helpers do not exist.

- [x] **Step 2: Implement typed cache helpers**

Create typed helpers in `internal/cache/article.go`, `category.go`, `contribution.go`, and `idempotency.go`. They should wrap the existing `Cache` interface and own key/TTL decisions.

- [x] **Step 3: Update public article list handler**

Modify `api/article.go` so `listArticle`:

1. Builds list params from query.
2. Calls `ArticleCache.GetList`.
3. Falls back to `ListArticles` and `CountArticles` on miss.
4. Calls `ArticleCache.SetList`.
5. Returns the same response shape as before.

- [x] **Step 4: Verify and commit**

Run:

```bash
gofmt -w internal/cache/*.go api/article.go api/article_test.go
go test ./internal/cache ./api
git add internal/cache/*.go api/article.go api/article_test.go
git commit -m "feat: cache public article list pages"
```

## Task 4: Invalidation Completeness and Contributions Fix

- [x] **Step 1: Write failing invalidation and contribution tests**

Add tests proving:

- Article update invalidation includes ID key, old slug key, new slug key, and list version bumps.
- Article delete invalidation includes ID key, slug key, category list key, and list version bump.
- Contributions cache stores raw upstream data and response slicing is bounded.

Run:

```bash
go test ./api ./gapi ./internal/cache
```

Expected: fail because invalidation and contribution behavior are not complete.

- [x] **Step 2: Implement invalidation updates**

Update article update/delete flows to fetch the previous article state before mutation and pass both previous and updated state into typed invalidation helpers. Do not bump article list versions for like/view increments.

- [x] **Step 3: Implement contribution cache fix**

Cache the raw GitHub contributions response. Slice only at response time and clamp indexes so short upstream data cannot panic.

- [x] **Step 4: Verify and commit**

Run:

```bash
gofmt -w api/user.go api/user_test.go gapi/rpc_update_article.go gapi/rpc_delete_article.go worker/task_delay_delete_cache.go
go test ./api ./gapi ./worker ./internal/cache
git add api gapi worker internal/cache
git commit -m "fix: complete cache invalidation semantics"
```

## Task 5: Singleflight, Docs, and Full Verification

- [x] **Step 1: Write failing singleflight tests**

Add tests for cache helpers or API handlers proving concurrent misses share one loader for article list and contributions where practical.

Run:

```bash
go test ./internal/cache ./api
```

Expected: fail before singleflight is added.

- [x] **Step 2: Implement singleflight**

Add process-local `singleflight.Group` use around article detail, article list, category list, and contributions read-through paths. Keep cache failures as fallback events rather than request failures.

- [x] **Step 3: Update environment docs**

Update `.env.example`, `docker-compose.yaml`, `docker-compose.dev.yaml`, and `README.md` with `REDIS_CACHE_DB` and `REDIS_QUEUE_DB`.

- [x] **Step 4: Full verification and commit**

Run:

```bash
gofmt -w .
go test -v -cover -short -count=1 ./...
git diff --check
git status --short
git add .
git commit -m "chore: document cache hardening configuration"
```

## Self-Review

- Spec coverage: covers Redis DB split, key naming, TTLs, typed helpers, public pagination cache, invalidation, contributions, singleflight, docs, and verification.
- Placeholder scan: no placeholder instructions remain; each task has concrete file paths and commands.
- Type consistency: plan consistently refers to `ArticleCache`, `CategoryCache`, `ContributionCache`, `IdempotencyCache`, `GetList`, `SetList`, and `BumpListVersion`.
