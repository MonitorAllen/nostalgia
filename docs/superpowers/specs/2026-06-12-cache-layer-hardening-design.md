# Cache Layer Hardening Design

## Background

Nostalgia currently uses Redis through a small `cache.Cache` interface and stores cache decisions directly in API/GAPI handlers. This works for current traffic, but key naming, TTL policy, cache invalidation, stampede protection, and Redis responsibility boundaries are scattered across the codebase.

This design strengthens the cache layer without replacing Redis or introducing a second Redis instance. Cache data and Asynq queue data will continue to share one Redis service, but use separate Redis DB indexes so cache eviction and queue reliability can be configured independently later.

## Goals

- Make cache keys explicit, collision-resistant, and grouped by namespace.
- Move TTL and jitter decisions out of handlers into typed cache helpers.
- Fix stale-cache risks around article ID keys, article slug keys, category list keys, and GitHub contributions.
- Add bounded cache support for public article pagination responses.
- Add singleflight protection for expensive read-through cache misses.
- Keep public reads resilient: cache errors should normally degrade to DB or upstream fetch.
- Keep idempotency behavior correct for likes and views.
- Split Redis cache and queue usage by DB index, not by separate Redis instances.

## Non-Goals

- Do not introduce a second Redis container or managed Redis instance.
- Do not add distributed locks for cache rebuilds in this phase.
- Do not redesign article, category, user, or queue business flows beyond cache boundaries.
- Do not replace Asynq.
- Do not add Prometheus or a metrics backend in this phase; structured logs are enough for the first pass.

## Redis DB Boundary

One Redis service remains in Docker Compose and local development.

- Cache DB: used by read-through cache and idempotency keys.
- Queue DB: used by Asynq.

Configuration will expose separate DB indexes:

- `REDIS_ADDRESS`: shared Redis host and port.
- `REDIS_CACHE_DB`: Redis DB index for cache use.
- `REDIS_QUEUE_DB`: Redis DB index for Asynq use.

Default values:

- `REDIS_CACHE_DB=0`
- `REDIS_QUEUE_DB=1`

The cache client must select `REDIS_CACHE_DB`. The Asynq Redis options must select `REDIS_QUEUE_DB`. Existing deployments that do not set these variables should keep working with safe defaults.

## Key Naming

Article cache keys will no longer share one ambiguous pattern.

- Article by ID: `cache:article:id:{uuid}`
- Article by slug: `cache:article:slug:{slug}`
- Article list version for all articles: `cache:article:list:version:all`
- Article list version for one category: `cache:article:list:version:category:{categoryID}`
- Article list page: `cache:article:list:v:{version}:category:{all|categoryID}:page:{page}:limit:{limit}`
- Category list: `cache:category:all`
- GitHub contributions: `cache:user:contributions`
- Article like by user: `idempotency:article:like:user:{articleID}:{userID}`
- Article like by guest: `idempotency:article:like:guest:{articleID}:{ip}`
- Article view by user: `idempotency:article:view:user:{articleID}:{userID}`
- Article view by guest: `idempotency:article:view:guest:{articleID}:{ip}`

The old `cache:article:{value}` pattern should be retired from runtime writes. During deployment, stale old keys can naturally expire or be ignored.

## TTL Policy

TTL decisions should live in cache-specific helpers instead of API handlers.

- Article by ID: 24 hours plus jitter.
- Article by slug: 24 hours plus jitter.
- Article list page: 15 minutes plus jitter.
- Empty article list page: 5 minutes plus jitter.
- Category list: 12 hours plus jitter.
- GitHub contributions: 12 hours plus jitter.
- Authenticated article like idempotency: 365 days.
- Guest article like idempotency: 7 days.
- Article view idempotency: 24 hours.

Jitter should be deterministic enough for tests and simple enough for production. A small randomized 5%-10% positive jitter is acceptable. Tests can use helper functions that assert a TTL falls within a known range instead of relying on exact duration.

## Typed Cache Layer

The low-level `cache.Cache` interface remains responsible for Redis primitives:

- `Ping`
- `Get`
- `Set`
- `Del`
- `SetNX`
- `TTL` or corrected expiration inspection
- `Close`

Business code should move toward typed cache helpers:

- `ArticleCache`
  - `GetByID`
  - `SetByID`
  - `GetBySlug`
  - `SetBySlug`
  - `GetList`
  - `SetList`
  - `BumpListVersion`
  - `Invalidate`
- `CategoryCache`
  - `GetList`
  - `SetList`
  - `InvalidateList`
- `ContributionCache`
  - `Get`
  - `Set`
- `IdempotencyCache`
  - `MarkArticleLikeOnce`
  - `MarkArticleViewOnce`

Handlers and RPC methods should call these helpers instead of building raw cache keys and TTL values directly.

## Article Pagination Cache

Public article pagination should cache the complete list response:

- `count`
- `articles`

The cache applies only to the public `/api/articles` list endpoint. Backend management lists and article search are not cached in this phase. Search has high-cardinality keyword input and should be designed separately.

The cache must be bounded:

- Cache only pages `1` through `5`.
- Include `limit` in the cache key.
- Treat `category_id=0` as the `all` category bucket.
- Use a shorter TTL than article detail cache because list responses include `views` and `likes`.

Article list cache invalidation should use versioned keys rather than enumerating page keys. The active version is stored separately for all articles and per category:

- `cache:article:list:version:all`
- `cache:article:list:version:category:{categoryID}`

The list response key includes the current version:

- `cache:article:list:v:{version}:category:{all|categoryID}:page:{page}:limit:{limit}`

When article writes affect public list membership or list display, the implementation increments the relevant version keys. Old page keys become unreachable and expire naturally.

## Invalidation Rules

Article update must invalidate:

- Article ID key for the article.
- New slug key when the article has a slug.
- Previous slug key when slug changes.
- Article list version for all articles when any field affecting public list output changes, including title, summary, cover, category, publish status, slug, read time, check-outdated state, and timestamps.
- Article list version for the old category and the new category when category membership may change.
- Category list key when category article counts may change.

Article delete must invalidate:

- Article ID key.
- Article slug key if present.
- Article list version for all articles.
- Article list version for the article category.
- Category list key.

Article list cache should not be invalidated for view or like increments. Those values may lag briefly in list responses to avoid making hot articles constantly invalidate pagination cache.

Category create, update, and delete must invalidate:

- Category list key.

GitHub contributions do not need explicit invalidation in this phase. They are TTL-driven.

The async cache deletion task must log deletion failures with the key name. It should continue processing the remaining keys when one delete fails.

## Contributions Cache Behavior

GitHub contributions should cache the raw upstream response, not the already sliced response. Each request can slice the cached raw data into the current response window. This avoids double-slicing on cache hits and keeps behavior stable across the day.

The response helper must guard against short or unexpected upstream data. If the current date is missing or there are fewer than 90 points after the current date, it should return a bounded slice instead of panicking.

## Singleflight

Use in-process `singleflight.Group` for read-through cache misses on:

- Article by ID.
- Article by slug.
- Article list page.
- Category list.
- GitHub contributions.

The first request that misses cache performs the DB or upstream call. Concurrent requests for the same cache key wait for that result. Cache errors should not prevent the DB/upstream fallback from running.

This is intentionally process-local. A distributed Redis lock is not needed for the current single-instance API deployment profile.

## Error Handling

Read-through cache paths:

- Cache get failure: log and fall back to DB/upstream.
- Cache set failure: log and return the fresh DB/upstream result.
- Unmarshal failure: treat as cache miss, log at error level, and overwrite cache after successful fallback.

Idempotency paths:

- `SetNX` failure remains a request error because it can affect like/view counting correctness.
- Duplicate idempotency key keeps returning conflict.

Invalidation paths:

- Admin write transactions should continue to use post-commit callbacks.
- Cache invalidation enqueue failure should surface as an error where the existing transaction flow expects the callback to succeed.
- Cache deletion task failures should be logged per key.

## Observability

Use structured logs with consistent fields:

- `cache_namespace`
- `cache_key`
- `cache_op`
- `cache_hit`
- `fallback`
- `module`

Initial implementation should use logs only. Metrics can be added later when the deployment has a metrics backend.

## Implementation Phases

### Phase 1: Redis DB Split and Cache Primitive Hardening

- Add config fields for `REDIS_CACHE_DB` and `REDIS_QUEUE_DB`.
- Make `NewRedisCache` select the cache DB.
- Make Asynq Redis options select the queue DB.
- Fix expiration inspection semantics or replace `IsExpired` with a clearer method.
- Add tests for Redis DB selection and TTL semantics.

### Phase 2: Key and TTL Semantics

- Update cache key constants to the new namespace format.
- Add TTL constants and jitter helpers.
- Update article, category, contributions, like, and view code to use the new key format.
- Add tests proving article ID and slug keys no longer collide.

### Phase 3: Typed Cache Helpers and Contributions Fix

- Introduce typed helpers for article detail, article list, category, contribution, and idempotency cache use.
- Move handler-level key and TTL decisions into these helpers.
- Cache bounded public article list pages with versioned page keys.
- Cache raw GitHub contributions and slice only at response time.
- Add tests for contribution cache hits and short upstream data.

### Phase 4: Invalidation Completeness

- Update article update/delete invalidation to include old slug, new slug, article ID, and category list keys.
- Update article update/delete invalidation to bump relevant article list versions.
- Keep category create/update/delete invalidating the category list.
- Improve delayed deletion task logging.
- Add tests for update/delete invalidation key sets.

### Phase 5: Singleflight and Final Verification

- Add process-local singleflight for expensive read-through paths.
- Add focused tests for duplicate concurrent loads where practical.
- Run full backend tests with `make test`.
- Run frontend tests only if frontend files are touched.
- Update documentation and `.env.example` for Redis DB configuration.

## Acceptance Criteria

- Runtime code no longer writes `cache:article:{value}` keys.
- Article ID and slug cache keys are distinct and covered by tests.
- Article cache TTLs are finite and include jitter.
- Public article list page cache is bounded to pages 1 through 5 and uses versioned keys.
- Article writes that affect public list output bump relevant list versions instead of scanning Redis keys.
- Contributions cache stores raw upstream data and does not double-slice on cache hits.
- Article slug updates invalidate the old slug key and the new slug key.
- Article deletion invalidates both ID and slug cache keys.
- Category-affecting article writes invalidate category list cache.
- Cache and queue use separate Redis DB indexes in configuration.
- Cache deletion task logs key-level delete failures.
- `make test` passes.

## Risks and Mitigations

- Risk: cache key migration leaves old Redis keys behind, including keys previously written without TTL.
  - Mitigation: new runtime ignores old keys; old keys can be cleared manually with a one-off Redis cleanup if needed.
- Risk: queue DB misconfiguration causes task loss or invisible queues.
  - Mitigation: default `REDIS_QUEUE_DB=1`, document clearly in `.env.example`, and test config loading.
- Risk: adding typed cache helpers creates too large a refactor.
  - Mitigation: migrate one surface at a time: article, category, contributions, idempotency.
- Risk: singleflight tests become brittle.
  - Mitigation: keep tests focused on helper behavior and avoid timing-sensitive assertions.
