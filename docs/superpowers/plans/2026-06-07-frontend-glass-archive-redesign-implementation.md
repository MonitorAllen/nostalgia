# Frontend Glass Archive Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild `web/frontend` as a PrimeVue-free glass archive blog UI with Tailwind CSS, owned Vue components, Bun-based frontend tooling, and persisted light/dark/system theme support.

**Architecture:** Replace PrimeVue with small owned components and CSS-token-driven Tailwind styles. Keep existing Vue Router, Pinia stores, API services, CKEditor, Prism, and CalHeatmap behavior while rewriting the app shell, article list, sidebar, results pages, article detail, auth forms, and feedback primitives. Theme resolution is isolated in a composable and applied through `data-theme`.

**Tech Stack:** Vue 3, Vite, Tailwind CSS, shadcn-vue/Reka-oriented owned components, Pinia, Axios, CKEditor, Prism, CalHeatmap.

---

## File Structure

- Modify `web/frontend/package.json`: remove PrimeVue packages, add Tailwind and UI helper dependencies, and run scripts through Bun.
- Modify `web/frontend/src/main.ts`: remove PrimeVue plugin setup and register no global PrimeVue directives.
- Create `web/frontend/tailwind.config.ts` and `web/frontend/postcss.config.cjs`: Tailwind setup.
- Replace `web/frontend/src/assets/main.css`: design tokens, Tailwind layers, glass archive surfaces, CKEditor content styling.
- Create `web/frontend/src/composables/useTheme.ts`: theme mode, resolved theme, localStorage persistence, system listener.
- Create `web/frontend/src/composables/useToast.ts`: small app-wide toast store.
- Create `web/frontend/src/components/ui/*.vue`: button, input, badge, panel, skeleton, toast, dialog, pagination, theme switcher, archive trail.
- Modify layout files under `web/frontend/src/views/layout/`: app shell, navigation, footer, sidebar.
- Modify article/category/search components and views under `web/frontend/src/components/` and `web/frontend/src/views/`.
- Modify auth views/components to remove PrimeVue.

## Task 1: Dependencies And Tooling

- [ ] Remove PrimeVue imports from package dependencies: `primevue`, `@primevue/themes`, `primeflex`, `primeicons`.
- [ ] Add Tailwind dependencies and utility helpers: `tailwindcss`, `tailwindcss-animate`, `class-variance-authority`, `clsx`, `tailwind-merge`, `@lucide/vue`, `reka-ui`.
- [ ] Add Tailwind and PostCSS config files.
- [ ] Verify `bun install` completes in `web/frontend`.

## Task 2: Theme Runtime And Global Styles

- [ ] Implement `useTheme.ts` with `system | light | dark` modes, `nostalgia-theme-mode` localStorage key, and `data-theme` application.
- [ ] Replace PrimeFlex/PrimeVue CSS imports with Tailwind layers and glass archive tokens.
- [ ] Define light and dark theme variables, including smoked graphite dark glass.
- [ ] Add base body, focus, scrollbar, form, reading, code, and reduced-motion styles.

## Task 3: Owned UI Primitives

- [ ] Build small owned UI components: button, input, badge, panel, skeleton, toast viewport, confirm dialog, pagination, theme switcher, and archive trail.
- [ ] Keep components typed and scoped to current frontend needs.
- [ ] Ensure keyboard focus and disabled states are visible.

## Task 4: Layout And Navigation

- [ ] Rewrite `App.vue` to render the app shell, toast viewport, navigation, route content, and footer without PrimeVue.
- [ ] Rewrite `NavBar.vue` with glass navigation, search, mobile menu, user menu, and theme switcher.
- [ ] Rewrite `MainLayout.vue` to use archive index layout with main column and sidebar glass panels.
- [ ] Rewrite `FooterView.vue` as quiet archive footer.

## Task 5: Home, Search, Category, And Sidebar

- [ ] Rewrite `ArticleList.vue` with archive list rows, loading skeletons, empty state, error state, stable image sizing, and owned pagination.
- [ ] Rewrite `CategoryList.vue` without PrimeVue DataView/Badge.
- [ ] Restyle `GithubContributions.vue` wrapper and heatmap legend for both themes.
- [ ] Replace PrimeVue Breadcrumb in search/category views with `ArchiveTrail`.

## Task 6: Article Detail And Comments

- [ ] Rewrite article detail template and styles into reading-first solid surfaces.
- [ ] Replace PrimeVue Divider, Button, Message, ConfirmDialog, tooltip, and toast usage.
- [ ] Keep article fetch, slug resolution, view increment, like increment, CKEditor comment input, Prism highlighting, and comment recursion behavior.
- [ ] Restyle copyright, like action, outdated warning, summary, comments, code, links, blockquotes, and reading progress.

## Task 7: Auth And Error Pages

- [ ] Rewrite login and register forms without PrimeVue.
- [ ] Keep existing store actions and navigation.
- [ ] Style forbidden/not-found/verify-email pages enough to match theme if they are simple.

## Task 8: Verification

- [ ] Run `rg "primevue|PrimeVue|primeflex|primeicons|@primevue" web/frontend/src web/frontend/package.json`.
- [ ] Run `bun run type-check` in `web/frontend`.
- [ ] Run `bun run build` in `web/frontend`.
- [ ] Start dev server and inspect desktop and mobile screenshots for home, article detail, search/category result, login/register, and theme modes.
- [ ] Verify theme persistence across reloads.
