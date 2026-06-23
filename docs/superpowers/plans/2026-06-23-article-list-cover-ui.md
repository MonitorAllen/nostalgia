# Article List Cover UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign public and admin article list cover presentation so covers share one rendering primitive while each list keeps its own reading or management purpose.

**Architecture:** Keep `ArticleCover` as the shared image primitive for fallback, aspect ratio, object-cover cropping, and image-error handling. Public and admin list views own layout and click semantics. Add source-contract tests for the shared rendering contract and admin preview behavior before changing templates.

**Tech Stack:** Vue 3, TypeScript, Tailwind CSS, Bun test, vue-tsc.

---

## File Structure

- Modify `web/frontend/src/components/article/articleCoverRendering.test.ts`
  - Extend source-contract tests for public list medium-emphasis layout and admin list cover behavior.
- Modify `web/frontend/src/components/article/ArticleList.vue`
  - Tune public reading-card layout, skeleton proportions, cover sizing, and content spacing.
- Modify `web/frontend/src/views/admin/AdminArticleListView.vue`
  - Replace manual cover `<img>` with `ArticleCover`.
  - Make cover click open preview instead of edit.
  - Tune row layout for desktop and mobile management use.

No backend, API, data model, or upload-policy files are part of this plan.

## Current Worktree Note

At the time this plan was written, the working tree also had local frontend edits in:

- `web/frontend/src/components/article/articleCoverRendering.test.ts`
- `web/frontend/src/util/seo.test.ts`
- `web/frontend/src/views/admin/AdminArticleEditorView.vue`

Do not discard those edits. When committing implementation work, stage only files that belong to the current task unless intentionally bundling verified earlier cover work.

---

### Task 1: Add List Cover Rendering Contract Tests

**Files:**
- Modify: `web/frontend/src/components/article/articleCoverRendering.test.ts`

- [ ] **Step 1: Add failing public and admin list cover tests**

Insert these tests inside the existing `describe('article cover rendering contracts', () => { ... })` block, after the current public `ArticleList` test:

```ts
  test('public article list presents covers as medium-emphasis reading cards', () => {
    const list = read('components/article/ArticleList.vue')

    expect(list).toContain("import ArticleCover from './ArticleCover.vue'")
    expect(list).toContain('md:grid-cols-[15rem_minmax(0,1fr)]')
    expect(list).toContain('variant="list"')
    expect(list).toContain('fallback-src="/images/go.png"')
    expect(list).toContain('md:aspect-[16/9]')
    expect(list).toContain('group-hover:border-accent/35')
    expect(list).not.toContain('md:grid-cols-[13rem_1fr]')
  })

  test('admin article list uses shared covers and previews from cover clicks', () => {
    const view = read('views/admin/AdminArticleListView.vue')

    expect(view).toContain("import ArticleCover from '@/components/article/ArticleCover.vue'")
    expect(view).toContain('<ArticleCover')
    expect(view).toContain(':src="article.cover"')
    expect(view).toContain('fallback-src="/images/go.png"')
    expect(view).toContain('@click="openArticlePreview(article)"')
    expect(view).toContain(":aria-label=\"`预览 ${article.title || '无标题文章'} 封面`\"")
    expect(view).not.toContain('coverLabel(article)')
    expect(view).not.toContain('@click="editArticle(article.id)"\\n            >\\n              <img')
  })
```

- [ ] **Step 2: Run the failing tests**

Run:

```bash
cd web/frontend
bun test src/components/article/articleCoverRendering.test.ts
```

Expected result before implementation: FAIL because `ArticleList.vue` still contains `md:grid-cols-[13rem_1fr]`, and `AdminArticleListView.vue` does not import or render `ArticleCover`.

- [ ] **Step 3: Commit the failing tests after implementation is ready**

Do not commit failing tests alone. Keep this task open until Tasks 2 and 3 make the tests pass.

---

### Task 2: Refine Public Article List Cover Layout

**Files:**
- Modify: `web/frontend/src/components/article/ArticleList.vue`

- [ ] **Step 1: Update loading skeleton proportions**

Replace the loading skeleton article wrapper:

```vue
<div v-for="i in 4" :key="i" class="archive-surface grid gap-4 rounded-archive p-4 md:grid-cols-[12rem_1fr]">
  <SkeletonBlock class="aspect-[16/9] w-full md:self-start" />
```

with:

```vue
<div
  v-for="i in 4"
  :key="i"
  class="archive-surface grid gap-4 rounded-archive p-3 md:grid-cols-[15rem_minmax(0,1fr)] md:p-4"
>
  <SkeletonBlock class="aspect-[16/9] w-full rounded-archive md:self-start" />
```

- [ ] **Step 2: Update public card layout classes**

Replace the public article wrapper class:

```vue
class="group archive-surface grid overflow-hidden rounded-archive transition duration-300 hover:-translate-y-0.5 hover:border-accent/35 md:grid-cols-[13rem_1fr]"
```

with:

```vue
class="group archive-surface grid overflow-hidden rounded-archive transition duration-300 hover:-translate-y-0.5 hover:border-accent/35 md:grid-cols-[15rem_minmax(0,1fr)] md:p-3"
```

- [ ] **Step 3: Update the public cover link frame**

Replace the cover `RouterLink` class:

```vue
class="relative block overflow-hidden bg-muted md:self-start"
```

with:

```vue
class="relative block overflow-hidden rounded-archive border border-border bg-muted transition-colors group-hover:border-accent/35 md:aspect-[16/9] md:self-start"
```

Keep the existing `ArticleCover` usage:

```vue
<ArticleCover
  :src="item.cover"
  :alt="item.title"
  variant="list"
  fallback-src="/images/go.png"
/>
```

- [ ] **Step 4: Tune public content spacing**

Replace the content wrapper class:

```vue
class="flex min-w-0 flex-col justify-between gap-4 p-4"
```

with:

```vue
class="flex min-w-0 flex-col justify-between gap-4 p-4 md:px-2 md:py-1"
```

Replace the title class:

```vue
class="m-0 text-xl font-black leading-snug text-foreground transition group-hover:text-accent md:text-2xl"
```

with:

```vue
class="m-0 text-xl font-black leading-snug text-foreground transition-colors group-hover:text-accent md:text-[1.45rem]"
```

- [ ] **Step 5: Run the public list contract test**

Run:

```bash
cd web/frontend
bun test src/components/article/articleCoverRendering.test.ts
```

Expected result at this point: the public list test passes; the admin list test still fails until Task 3 is complete.

---

### Task 3: Refactor Admin Article List Covers

**Files:**
- Modify: `web/frontend/src/views/admin/AdminArticleListView.vue`

- [ ] **Step 1: Import the shared cover component**

Add this import near the other component imports:

```ts
import ArticleCover from '@/components/article/ArticleCover.vue'
```

- [ ] **Step 2: Remove the manual cover label helper**

Delete this helper:

```ts
const coverLabel = (article: AdminArticle) => {
  return article.cover?.trim() || '/images/go.png'
}
```

`ArticleCover` will receive `article.cover` and `fallback-src="/images/go.png"` directly.

- [ ] **Step 3: Replace the admin cover button**

Replace the existing cover button block:

```vue
<button
  type="button"
  class="group hidden h-28 w-40 shrink-0 overflow-hidden rounded-archive border border-border bg-muted focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent sm:block"
  @click="editArticle(article.id)"
>
  <img
    :src="coverLabel(article)"
    :alt="article.title ? `${article.title} 封面` : '文章封面'"
    class="h-full w-full object-cover transition duration-200 group-hover:scale-[1.03]"
    loading="lazy"
  />
</button>
```

with:

```vue
<button
  type="button"
  class="group block w-full shrink-0 overflow-hidden rounded-archive border border-border bg-muted transition-colors hover:border-accent/35 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent sm:w-44 lg:w-40"
  :aria-label="`预览 ${article.title || '无标题文章'} 封面`"
  @click="openArticlePreview(article)"
>
  <ArticleCover
    :src="article.cover"
    :alt="article.title ? `${article.title} 封面` : '文章封面'"
    variant="list"
    fallback-src="/images/go.png"
  />
</button>
```

- [ ] **Step 4: Tune admin row layout for mobile and desktop**

Replace the row body wrapper:

```vue
<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
  <div class="flex min-w-0 flex-1 gap-4">
```

with:

```vue
<div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
  <div class="flex min-w-0 flex-1 flex-col gap-4 sm:flex-row">
```

Replace the content wrapper:

```vue
<div class="min-w-0 flex-1 space-y-3">
```

with:

```vue
<div class="min-w-0 flex-1 space-y-3 sm:pt-0.5">
```

- [ ] **Step 5: Keep title preview behavior unchanged**

Verify the title button still uses:

```vue
@click="openArticlePreview(article)"
```

and still wraps the `h2` title. Do not change title clicks to edit.

- [ ] **Step 6: Run the list cover contract tests**

Run:

```bash
cd web/frontend
bun test src/components/article/articleCoverRendering.test.ts
```

Expected result: PASS for all article cover rendering contract tests.

- [ ] **Step 7: Commit public and admin list cover changes**

Stage only the files changed in Tasks 1-3:

```bash
git add web/frontend/src/components/article/articleCoverRendering.test.ts \
  web/frontend/src/components/article/ArticleList.vue \
  web/frontend/src/views/admin/AdminArticleListView.vue
git commit -m "feat(frontend): refine article list cover UI"
```

Expected result: commit succeeds. If `articleCoverRendering.test.ts` also contains verified earlier cover diagnostics edits, include them only if they are intentional for this branch.

---

### Task 4: Full Frontend Verification

**Files:**
- No source edits expected.

- [ ] **Step 1: Run focused cover and SEO tests**

Run:

```bash
cd web/frontend
bun test src/components/article/articleCoverPolicy.test.ts src/components/article/articleCoverRendering.test.ts src/util/seo.test.ts
```

Expected result: PASS.

- [ ] **Step 2: Run the full frontend test suite**

Run:

```bash
cd web/frontend
bun test
```

Expected result: PASS.

- [ ] **Step 3: Run type checking**

Run:

```bash
cd web/frontend
bun run type-check
```

Expected result: command exits with code 0.

- [ ] **Step 4: Run production build**

Run:

```bash
cd web/frontend
bun run build-only
```

Expected result: command exits with code 0 and produces the Vite build output.

- [ ] **Step 5: Inspect git status**

Run:

```bash
git status --short
```

Expected result: only intentional, already-reviewed files remain modified or the working tree is clean.

If verification required no additional edits, do not create a verification-only commit.
