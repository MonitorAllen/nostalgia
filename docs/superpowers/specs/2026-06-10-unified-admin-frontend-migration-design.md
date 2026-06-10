# Unified Admin Frontend Migration Design

## Goal

Move the owner-only blog administration experience from `web/backend` into the redesigned `web/frontend` app under `/admin`, while keeping the public blog reading experience and admin writing experience visually consistent.

## Current Baseline

- `master` includes PR #30, so the public frontend already uses the glass/archive visual direction.
- `web/frontend` uses Vue 3, Vite, TypeScript, Tailwind CSS, Reka UI, lucide icons, Pinia, CKEditor, DOMPurify, custom toast components, and Bun scripts.
- PrimeVue has been removed from the frontend.
- Public frontend API calls use `VITE_APP_BASE_URL=/api` and the Vite dev server proxies `/api` to `http://localhost:8080`.
- Old admin functionality in `web/backend` uses Geeker Admin, Element Plus, a separate axios layer, and gRPC-Gateway endpoints exposed under `/v1`.
- Old admin article authoring contains useful business logic: article list, create/update/delete, categories, CKEditor config, upload adapter, draft cache, dirty-state tracking, cover upload, and save status.

## Non-Goals

- Do not preserve Geeker Admin, Element Plus, dynamic admin menus, role management, department management, dictionary management, logs, timing tasks, or broad RBAC screens.
- Do not create a second frontend app or a separate admin visual language.
- Do not merge public user auth and admin auth into one storage namespace.
- Do not refactor backend APIs unless a frontend integration issue makes it necessary.

## Architecture

The project will become a single frontend app:

- Public blog routes remain under `/`.
- Owner admin routes are added under `/admin`.
- `/admin` uses `AdminLayout`, which hides public navigation and footer but reuses the same design tokens, glass surfaces, typography, theme mode, toast viewport, confirm dialog, and icon system.
- Admin pages are implemented with small local components inside `web/frontend/src/views/admin` and `web/frontend/src/components/admin`.

The old `web/backend` directory remains in the repository until the new `/admin` core workflows are verified. After verification, it will be deleted in a cleanup phase.

## Routing

New route group:

- `/admin/login`: admin login page.
- `/admin`: redirects to `/admin/articles`.
- `/admin/articles`: article management list.
- `/admin/articles/new`: creates a draft and routes to the editor when an id is available.
- `/admin/articles/:id/edit`: article editor.
- `/admin/categories`: category management.

Route guard behavior:

- Routes with `meta.requiresAdmin` require a valid admin access token.
- If unauthenticated, redirect to `/admin/login?redirect=<current path>`.
- If an admin token refresh fails, clear admin auth and redirect to `/admin/login`.
- Public user login at `/login` is not used for admin access.

## API Design

Keep public and admin API clients separate:

- Public API client continues to use `/api` through `web/frontend/src/util/http.ts`.
- Add `web/frontend/src/admin/api/adminHttp.ts` for `/v1` gRPC-Gateway calls.
- Add Vite proxy entry `/v1 -> http://localhost:9091/v1`.
- Production deployment should route `/v1` to the gRPC-Gateway service, matching the dev URL shape.

Admin API modules:

- `adminAuthApi.ts`: `POST /admin/login`, `POST /admin/renew_access`, optional `GET /admin/info`.
- `adminArticleApi.ts`: `GET /articles`, `GET /articles/{id}/{need_content}`, `POST /articles`, `PATCH /articles`, `DELETE /articles/{id}`.
- `adminCategoryApi.ts`: `GET /categories/all`, `POST /categories`, `PATCH /categories`, `DELETE /categories/{id}`.
- `adminUploadApi.ts`: `POST /util/upload_file`.

All admin modules call `adminHttp`, which adds `Authorization: Bearer <admin token>` unless `skipAuth` is set.

## Admin Auth

Create a dedicated Pinia store:

- File: `web/frontend/src/admin/stores/adminAuth.ts`.
- Local storage keys use a separate `nostalgia_admin_*` prefix.
- Stored values: admin info, access token, access token expiry, refresh token, refresh token expiry.
- Exposed state: `isAuthenticated`, `token`, `admin`.
- Exposed actions: `login`, `logout`, `refreshAccessToken`, `setTokens`, `clear`.

This avoids collisions with public user auth and keeps admin sessions independent from visitor/comment accounts.

## Visual Design

Admin pages should feel like a quiet writing workspace, not an enterprise dashboard:

- Use the existing glass/archive theme foundation.
- Prefer dense but calm layouts: compact top bars, clear tables/lists, restrained panels, and direct actions.
- Use lucide icons for icon buttons.
- Use existing `AppButton`, `AppInput`, `AppBadge`, `ArchivePanel`, `ConfirmDialog`, `ToastViewport`, and `ThemeSwitcher`.
- Avoid nested cards, oversized hero sections, marketing copy, and generic dashboard widgets.
- Keep dark mode neutral charcoal/green-gray like the current frontend instead of returning to a heavy blue admin theme.

## Admin Layout

`AdminLayout` includes:

- Left navigation on desktop with links for articles and categories.
- Compact top bar with current section title, theme switcher, "view site" link, and logout button.
- Mobile top navigation with a collapsible menu.
- Main content area constrained for readable management screens.

The layout should not show public blog nav/footer or public account links.

## Article Management

Article list requirements:

- Search by title when supported by the existing admin list endpoint.
- Paginated or incremental loading with `page` and `limit`.
- Show title, summary excerpt, category, created/updated time, views, likes, and publish state.
- Actions: edit, toggle publish/draft, delete.
- Delete uses the shared confirm dialog.
- Status changes and deletes use toast feedback.

Article editor requirements:

- Preserve old admin business behavior: create draft first, edit title/summary/slug/category/cover/publish state/content, save changes, track dirty state, and cache unsaved draft in `sessionStorage`.
- Use CKEditor in the frontend app.
- Use `/v1/util/upload_file` for content image and cover uploads.
- Show character and word count.
- Warn when leaving with unsaved changes.
- Provide a focused full-screen or wide writing mode if it does not complicate the first migration.

## Reading and Writing Style Consistency

The editor content area should closely match frontend article rendering:

- Use `reading-prose` or an admin-specific class derived from `reading-prose` for CKEditor content.
- Keep heading scale, paragraph rhythm, blockquote, code block, table, image, and figcaption styles aligned with `web/frontend/src/assets/content.css`.
- Render saved article content through the same sanitized reading styles on the public article page.
- Avoid admin-only content styles that make the article look different after publish.

## Category Management

Category page requirements:

- List all categories using `/v1/categories/all`.
- Show name, article count, created time, and updated time when available.
- Create category.
- Rename category.
- Delete category with confirmation.
- Use small inline forms or lightweight dialogs rather than a full enterprise table framework.

## Cleanup Strategy

After `/admin` login, article list, article editor, category management, and frontend build are verified:

- Delete `web/backend`.
- Update root `AGENTS.md` to remove the old backend frontend instructions and document the unified admin route.
- Update any deployment or build docs that still mention `web/backend`.
- Keep backend gRPC-Gateway code because `/v1` remains the admin API surface.

## Commit Strategy

Use several focused commits:

1. `docs: add unified admin frontend migration design`
2. `feat(frontend): add admin api client and auth store`
3. `feat(frontend): add admin layout and route guard`
4. `feat(frontend): add admin article management`
5. `feat(frontend): add admin article editor`
6. `feat(frontend): add admin category management`
7. `chore(frontend): remove legacy backend admin app`

The exact split can be adjusted during implementation, but large unrelated changes should not be collapsed into one commit.

## Verification

For frontend-only phases:

```bash
cd web/frontend
bun run type-check
bun run build
```

For phases that touch backend routing or generated code:

```bash
make test
```

Manual verification should cover:

- Public blog still renders articles and comments.
- `/admin/login` logs in with admin credentials.
- `/admin/articles` lists drafts and published articles.
- Creating a new article produces an editable draft.
- CKEditor content upload works through `/v1/util/upload_file`.
- Saving content preserves reading style on the public article page.
- Category create, rename, and delete work.
- Theme mode persists across public and admin pages.
- Admin auth expiration refreshes or redirects cleanly.

## Open Confirmation

- Production reverse proxy must expose `/v1` to the gRPC-Gateway service.
- The first migration keeps only owner-needed admin pages; broader system-management pages remain intentionally removed.
