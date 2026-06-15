# SEO/GEO Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add technical SEO/GEO foundations for public Nostalgia articles.

**Architecture:** Keep the existing Vue SPA and Go API. Vue owns runtime metadata and JSON-LD, while Go serves crawler discovery files from live published content.

**Tech Stack:** Go/Gin/sqlc, Vue 3/Vite/TypeScript, Bun tests, Nginx, Docker Compose.

---

### Task 1: Frontend Metadata Utility

**Files:**
- Create: `web/frontend/src/util/seo.ts`
- Create: `web/frontend/src/util/seo.test.ts`

- [ ] Write failing Bun tests for origin normalization, plain text extraction, URL building, Article JSON-LD generation, and managed DOM tag upserts.
- [ ] Run `cd web/frontend && bun test src/util/seo.test.ts` and verify the tests fail because `seo.ts` does not exist.
- [ ] Implement the minimal metadata utility.
- [ ] Run `cd web/frontend && bun test src/util/seo.test.ts` and verify it passes.
- [ ] Commit with `feat(frontend): add seo metadata utility`.

### Task 2: Frontend Route and Article Metadata

**Files:**
- Modify: `web/frontend/index.html`
- Modify: `web/frontend/src/router/index.ts`
- Modify: `web/frontend/src/views/article/ArticleView.vue`
- Modify: `web/frontend/src/views/HomeView.vue`
- Modify: `web/frontend/src/views/category/CategoryArticleView.vue`
- Modify: `web/frontend/src/views/article/SearchArticleView.vue`

- [ ] Add failing tests that assert static route metadata defaults and article metadata builders.
- [ ] Run targeted Bun tests and verify they fail.
- [ ] Wire route metadata and article metadata updates through the SEO utility.
- [ ] Run targeted Bun tests and `cd web/frontend && bun run type-check`.
- [ ] Commit with `feat(frontend): wire seo metadata into public routes`.

### Task 3: Backend Robots and Sitemap

**Files:**
- Create: `api/seo.go`
- Create: `api/seo_test.go`
- Modify: `api/server.go`
- Modify: `db/query/article.sql`
- Regenerate: `db/sqlc/*`, `db/mock/store.go`

- [ ] Add failing API tests for `GET /robots.txt` and `GET /sitemap.xml`.
- [ ] Add a published sitemap article query and run `make sqlc && make mock`.
- [ ] Implement robots and sitemap handlers.
- [ ] Run `make test`.
- [ ] Commit with `feat(api): serve robots and sitemap`.

### Task 4: Deployment and Documentation

**Files:**
- Modify: `web/nginx.conf`
- Modify: `web/frontend/src/deploy/nginxConfig.test.ts`
- Modify: `.env.example`
- Modify: `.github/workflows/test.yml`
- Modify: `README.md`

- [ ] Add failing deployment tests for `/robots.txt` and `/sitemap.xml` proxying.
- [ ] Update Nginx to proxy crawler discovery files to the API.
- [ ] Document `DOMAIN` as the canonical public origin and explain SEO/GEO behavior.
- [ ] Run frontend deploy tests and Compose config checks.
- [ ] Commit with `docs: document seo geo configuration`.

### Task 5: Final Verification

- [ ] Run `make test`.
- [ ] Run `cd web/frontend && bun test`.
- [ ] Run `cd web/frontend && bun run type-check`.
- [ ] Run `cd web/frontend && bun run build`.
- [ ] Run `docker compose config --quiet`.
- [ ] Run `docker compose -f docker-compose.dev.yaml config --quiet`.
- [ ] Run `git diff --check`.
- [ ] Commit any final docs-only fixes if needed.
