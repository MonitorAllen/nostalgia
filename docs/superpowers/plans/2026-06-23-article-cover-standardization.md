# Article Cover Standardization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Standardize article cover rendering and guidance across public article lists, public article detail pages, backend previews, and social metadata without adding backend image processing or new storage fields.

**Architecture:** Keep the existing `article.cover` data model. Add a small pure cover policy module for dimensions, ratio checks, and warnings; add a reusable cover display component for stable `16:9` rendering; and make the backend editor consume the same policy through a presentational cover panel.

**Tech Stack:** Vue 3, TypeScript, Tailwind CSS, Bun test, existing source-contract test style.

---

## File Structure

- Create `web/frontend/src/components/article/articleCoverPolicy.ts`
  - Owns cover ratio constants, recommended dimensions, warning classification, and browser image dimension loading helpers.
- Create `web/frontend/src/components/article/articleCoverPolicy.test.ts`
  - Tests pure policy behavior with Bun.
- Create `web/frontend/src/components/article/ArticleCover.vue`
  - Renders a stable cover container with `object-cover` and fallback behavior.
- Create `web/frontend/src/components/article/articleCoverRendering.test.ts`
  - Uses source-contract checks for public reader/list integration and backend panel wiring.
- Create `web/frontend/src/admin/editor/AdminArticleCoverPanel.vue`
  - Displays upload/remove actions, guidance text, warnings, and detail/list/share previews.
- Modify `web/frontend/src/components/article/ArticleReader.vue`
  - Moves cover rendering above the title/header and uses `ArticleCover`.
- Modify `web/frontend/src/components/article/ArticleList.vue`
  - Uses `ArticleCover` in list cards and removes `object-contain` padded thumbnails.
- Modify `web/frontend/src/views/article/ArticleView.vue`
  - Passes `article.cover` and `show-cover` into `ArticleReader`.
- Modify `web/frontend/src/views/admin/AdminArticleEditorView.vue`
  - Replaces inline cover markup with `AdminArticleCoverPanel` and wires non-blocking cover diagnostics.
- Modify `web/frontend/src/util/seo.test.ts`
  - Adds an explicit assertion that Twitter large image behavior remains tied to cover presence.

---

### Task 1: Add Cover Policy

**Files:**
- Create: `web/frontend/src/components/article/articleCoverPolicy.ts`
- Create: `web/frontend/src/components/article/articleCoverPolicy.test.ts`

- [ ] **Step 1: Write the failing policy tests**

Create `web/frontend/src/components/article/articleCoverPolicy.test.ts`:

```typescript
import { describe, expect, test } from 'bun:test'
import {
  ARTICLE_COVER_ASPECT_RATIO,
  ARTICLE_COVER_MIN_HEIGHT,
  ARTICLE_COVER_MIN_WIDTH,
  ARTICLE_COVER_RECOMMENDED_HEIGHT,
  ARTICLE_COVER_RECOMMENDED_WIDTH,
  inspectArticleCoverDimensions
} from './articleCoverPolicy'

describe('article cover policy', () => {
  test('defines the canonical 16:9 cover standard', () => {
    expect(ARTICLE_COVER_ASPECT_RATIO).toBe(16 / 9)
    expect(ARTICLE_COVER_RECOMMENDED_WIDTH).toBe(1600)
    expect(ARTICLE_COVER_RECOMMENDED_HEIGHT).toBe(900)
    expect(ARTICLE_COVER_MIN_WIDTH).toBe(1200)
    expect(ARTICLE_COVER_MIN_HEIGHT).toBe(675)
  })

  test('accepts recommended 16:9 cover dimensions without warnings', () => {
    expect(inspectArticleCoverDimensions({ width: 1600, height: 900 })).toEqual({
      width: 1600,
      height: 900,
      ratio: 16 / 9,
      status: 'ok',
      warnings: []
    })
  })

  test('warns when the image is below minimum recommended dimensions', () => {
    const result = inspectArticleCoverDimensions({ width: 900, height: 506 })

    expect(result.status).toBe('warning')
    expect(result.warnings).toContain('建议封面至少为 1200x675，当前图片可能在高清屏上显得模糊。')
  })

  test('warns when the image ratio is far from 16:9', () => {
    const result = inspectArticleCoverDimensions({ width: 1200, height: 1200 })

    expect(result.status).toBe('warning')
    expect(result.warnings).toContain('当前图片比例偏离 16:9，详情页、列表或分享预览中可能出现明显裁切。')
  })

  test('warns when dimensions cannot be read', () => {
    expect(inspectArticleCoverDimensions(null)).toEqual({
      width: 0,
      height: 0,
      ratio: 0,
      status: 'warning',
      warnings: ['无法读取图片尺寸，仍可继续保存。']
    })
  })
})
```

- [ ] **Step 2: Run policy tests and verify they fail**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverPolicy.test.ts
```

Expected: the test command fails because `articleCoverPolicy.ts` does not exist.

- [ ] **Step 3: Add the policy implementation**

Create `web/frontend/src/components/article/articleCoverPolicy.ts`:

```typescript
export const ARTICLE_COVER_ASPECT_RATIO = 16 / 9
export const ARTICLE_COVER_RATIO_TOLERANCE = 0.12
export const ARTICLE_COVER_RECOMMENDED_WIDTH = 1600
export const ARTICLE_COVER_RECOMMENDED_HEIGHT = 900
export const ARTICLE_COVER_HIGH_RES_WIDTH = 1920
export const ARTICLE_COVER_HIGH_RES_HEIGHT = 1080
export const ARTICLE_COVER_MIN_WIDTH = 1200
export const ARTICLE_COVER_MIN_HEIGHT = 675

export type ArticleCoverInspectionStatus = 'ok' | 'warning'

export interface ArticleCoverDimensions {
  width: number
  height: number
}

export interface ArticleCoverInspection {
  width: number
  height: number
  ratio: number
  status: ArticleCoverInspectionStatus
  warnings: string[]
}

const LOW_RESOLUTION_WARNING =
  '建议封面至少为 1200x675，当前图片可能在高清屏上显得模糊。'
const OFF_RATIO_WARNING = '当前图片比例偏离 16:9，详情页、列表或分享预览中可能出现明显裁切。'
const UNREADABLE_WARNING = '无法读取图片尺寸，仍可继续保存。'

export function inspectArticleCoverDimensions(
  dimensions: ArticleCoverDimensions | null
): ArticleCoverInspection {
  if (!dimensions || dimensions.width <= 0 || dimensions.height <= 0) {
    return {
      width: 0,
      height: 0,
      ratio: 0,
      status: 'warning',
      warnings: [UNREADABLE_WARNING]
    }
  }

  const ratio = dimensions.width / dimensions.height
  const warnings: string[] = []

  if (dimensions.width < ARTICLE_COVER_MIN_WIDTH || dimensions.height < ARTICLE_COVER_MIN_HEIGHT) {
    warnings.push(LOW_RESOLUTION_WARNING)
  }

  if (Math.abs(ratio - ARTICLE_COVER_ASPECT_RATIO) > ARTICLE_COVER_RATIO_TOLERANCE) {
    warnings.push(OFF_RATIO_WARNING)
  }

  return {
    width: dimensions.width,
    height: dimensions.height,
    ratio,
    status: warnings.length > 0 ? 'warning' : 'ok',
    warnings
  }
}

export function loadArticleCoverDimensions(src: string): Promise<ArticleCoverDimensions> {
  return new Promise((resolve, reject) => {
    if (!src) {
      reject(new Error(UNREADABLE_WARNING))
      return
    }

    const image = new Image()
    image.onload = () => resolve({ width: image.naturalWidth, height: image.naturalHeight })
    image.onerror = () => reject(new Error(UNREADABLE_WARNING))
    image.src = src
  })
}

export function loadArticleCoverFileDimensions(file: File): Promise<ArticleCoverDimensions> {
  return new Promise((resolve, reject) => {
    const objectUrl = URL.createObjectURL(file)

    loadArticleCoverDimensions(objectUrl)
      .then(resolve)
      .catch(reject)
      .finally(() => URL.revokeObjectURL(objectUrl))
  })
}
```

- [ ] **Step 4: Run policy tests and verify they pass**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverPolicy.test.ts
```

Expected: all tests in `articleCoverPolicy.test.ts` pass.

- [ ] **Step 5: Commit the policy module**

Run:

```bash
git add web/frontend/src/components/article/articleCoverPolicy.ts web/frontend/src/components/article/articleCoverPolicy.test.ts
git commit -m "feat(frontend): add article cover policy"
```

---

### Task 2: Add Shared Cover Rendering Component

**Files:**
- Create: `web/frontend/src/components/article/ArticleCover.vue`
- Create: `web/frontend/src/components/article/articleCoverRendering.test.ts`

- [ ] **Step 1: Write source-contract tests for shared rendering**

Create `web/frontend/src/components/article/articleCoverRendering.test.ts`:

```typescript
import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const src = (...parts: string[]) => resolve(import.meta.dir, '..', '..', ...parts)
const read = (...parts: string[]) => readFileSync(src(...parts), 'utf8')

describe('article cover rendering contracts', () => {
  test('shared ArticleCover component uses a stable 16:9 object-cover container', () => {
    const component = read('components/article/ArticleCover.vue')

    expect(component).toContain('aspect-[16/9]')
    expect(component).toContain('object-cover')
    expect(component).toContain('fallbackSrc')
    expect(component).toContain('handleImageError')
  })
})
```

- [ ] **Step 2: Run rendering contract tests and verify they fail**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverRendering.test.ts
```

Expected: the test command fails because `ArticleCover.vue` does not exist.

- [ ] **Step 3: Add the shared cover component**

Create `web/frontend/src/components/article/ArticleCover.vue`:

```vue
<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    src?: string
    alt?: string
    variant?: 'detail' | 'list' | 'preview'
    fallbackSrc?: string
  }>(),
  {
    src: '',
    alt: '文章封面',
    variant: 'detail',
    fallbackSrc: ''
  }
)

const normalizeSrc = (src: string) => src.trim()
const resolveDisplaySrc = (src: string) => normalizeSrc(src) || props.fallbackSrc
const displaySrc = ref(resolveDisplaySrc(props.src))

watch(
  () => [props.src, props.fallbackSrc] as const,
  ([nextSrc]) => {
    displaySrc.value = resolveDisplaySrc(nextSrc)
  }
)

const containerClass = computed(() => {
  const base = 'relative block aspect-[16/9] w-full overflow-hidden bg-muted'
  if (props.variant === 'list') return `${base} h-full`
  if (props.variant === 'preview') return `${base} rounded-archive`
  return `${base} rounded-archive`
})

const imageClass = computed(() => {
  const base = 'h-full w-full object-cover transition duration-500'
  if (props.variant === 'list') return `${base} group-hover:scale-[1.03]`
  return base
})

const handleImageError = () => {
  if (props.fallbackSrc && displaySrc.value !== props.fallbackSrc) {
    displaySrc.value = props.fallbackSrc
    return
  }

  displaySrc.value = ''
}
</script>

<template>
  <span v-if="displaySrc" :class="containerClass">
    <img :src="displaySrc" :alt="alt" :class="imageClass" @error="handleImageError" />
  </span>
</template>
```

- [ ] **Step 4: Run component contract tests**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverRendering.test.ts
```

Expected: `shared ArticleCover component uses a stable 16:9 object-cover container` passes.

- [ ] **Step 5: Commit the shared component**

Run:

```bash
git add web/frontend/src/components/article/ArticleCover.vue web/frontend/src/components/article/articleCoverRendering.test.ts
git commit -m "feat(frontend): add shared article cover component"
```

---

### Task 3: Wire Public Detail And List Rendering

**Files:**
- Modify: `web/frontend/src/components/article/ArticleReader.vue`
- Modify: `web/frontend/src/views/article/ArticleView.vue`
- Modify: `web/frontend/src/components/article/ArticleList.vue`
- Modify: `web/frontend/src/components/article/articleCoverRendering.test.ts`

- [ ] **Step 1: Extend rendering contract tests**

Append these tests to `web/frontend/src/components/article/articleCoverRendering.test.ts`:

```typescript
test('ArticleReader renders the cover before the header when showCover is enabled', () => {
  const reader = read('components/article/ArticleReader.vue')

  expect(reader).toContain("import ArticleCover from './ArticleCover.vue'")
  expect(reader).toContain('<ArticleCover')
  expect(reader.indexOf('<ArticleCover')).toBeLessThan(reader.indexOf('<header'))
  expect(reader).toContain(':src="cover"')
  expect(reader).toContain('variant="detail"')
})

test('public article detail passes cover into ArticleReader', () => {
  const view = read('views/article/ArticleView.vue')

  expect(view).toContain(':cover="article.cover"')
  expect(view).toContain('show-cover')
})

test('ArticleList uses shared object-cover rendering without padded contain thumbnails', () => {
  const list = read('components/article/ArticleList.vue')

  expect(list).toContain("import ArticleCover from './ArticleCover.vue'")
  expect(list).toContain('variant="list"')
  expect(list).not.toContain('object-contain')
  expect(list).not.toContain('p-2 transition')
})
```

- [ ] **Step 2: Run rendering tests and verify new tests fail**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverRendering.test.ts
```

Expected: the new tests fail because public surfaces still use the old rendering.

- [ ] **Step 3: Update ArticleReader**

Modify `web/frontend/src/components/article/ArticleReader.vue`.

Add the import:

```typescript
import ArticleCover from './ArticleCover.vue'
```

Place this block immediately inside `<article>` and before `<header>`:

```vue
<ArticleCover
  v-if="showCover && cover"
  :src="cover"
  :alt="displayTitle || '文章封面'"
  variant="detail"
  class="mb-6"
/>
```

Remove the existing lower cover `<img>` block that currently appears after the summary section:

```vue
<img
  v-if="showCover && cover"
  :src="cover"
  :alt="displayTitle || '文章封面'"
  class="mb-6 aspect-[16/9] w-full rounded-archive object-cover"
/>
```

- [ ] **Step 4: Update public article detail**

Modify the `ArticleReader` call in `web/frontend/src/views/article/ArticleView.vue` to pass cover data:

```vue
<ArticleReader
  v-if="article"
  :title="article.title"
  :summary="article.summary"
  :content="article.content"
  :category-name="article.category_name"
  :read-time="article.read_time"
  :cover="article.cover"
  :created-at="article.created_at"
  :likes="article.likes"
  :views="article.views"
  show-cover
/>
```

- [ ] **Step 5: Update public article list**

Modify `web/frontend/src/components/article/ArticleList.vue`.

Add the import:

```typescript
import ArticleCover from './ArticleCover.vue'
```

Remove the `onImageError` function because fallback behavior moves into `ArticleCover`.

Replace the image `RouterLink` with:

```vue
<RouterLink
  :to="`/article/${item.slug ? item.slug : item.id}`"
  class="relative block aspect-[16/9] overflow-hidden bg-muted md:h-full"
>
  <ArticleCover
    :src="item.cover"
    :alt="item.title"
    variant="list"
    fallback-src="/images/go.png"
  />
</RouterLink>
```

Update the loading skeleton image block from:

```vue
<SkeletonBlock class="h-40 w-full md:h-32" />
```

to:

```vue
<SkeletonBlock class="aspect-[16/9] w-full md:h-full" />
```

- [ ] **Step 6: Run rendering tests**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverRendering.test.ts
```

Expected: all tests in `articleCoverRendering.test.ts` pass.

- [ ] **Step 7: Commit public rendering changes**

Run:

```bash
git add web/frontend/src/components/article/ArticleReader.vue web/frontend/src/views/article/ArticleView.vue web/frontend/src/components/article/ArticleList.vue web/frontend/src/components/article/articleCoverRendering.test.ts
git commit -m "feat(frontend): standardize public article covers"
```

---

### Task 4: Add Backend Cover Preview Panel

**Files:**
- Create: `web/frontend/src/admin/editor/AdminArticleCoverPanel.vue`
- Modify: `web/frontend/src/components/article/articleCoverRendering.test.ts`

- [ ] **Step 1: Add source-contract tests for backend panel**

Append this test to `web/frontend/src/components/article/articleCoverRendering.test.ts`:

```typescript
test('admin cover panel exposes guidance, warnings, and multi-surface previews', () => {
  const panel = read('admin/editor/AdminArticleCoverPanel.vue')

  expect(panel).toContain('推荐 1600x900')
  expect(panel).toContain('1920x1080')
  expect(panel).toContain('详情页头图')
  expect(panel).toContain('列表卡片')
  expect(panel).toContain('分享预览')
  expect(panel).toContain('inspection.warnings')
  expect(panel).toContain('ArticleCover')
})
```

- [ ] **Step 2: Run rendering tests and verify the panel test fails**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverRendering.test.ts
```

Expected: the new backend panel test fails because `AdminArticleCoverPanel.vue` does not exist.

- [ ] **Step 3: Add the backend cover panel**

Create `web/frontend/src/admin/editor/AdminArticleCoverPanel.vue`:

```vue
<script setup lang="ts">
import { ImagePlus, Trash2 } from '@lucide/vue'
import ArticleCover from '@/components/article/ArticleCover.vue'
import AppButton from '@/components/ui/AppButton.vue'
import type { ArticleCoverInspection } from '@/components/article/articleCoverPolicy'

withDefaults(
  defineProps<{
    cover?: string
    title?: string
    isUploading?: boolean
    inspection?: ArticleCoverInspection | null
  }>(),
  {
    cover: '',
    title: '',
    isUploading: false,
    inspection: null
  }
)

defineEmits<{
  upload: []
  remove: []
}>()
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center justify-between gap-3">
      <div>
        <span class="block text-sm font-bold text-foreground">封面图</span>
        <span class="block text-xs leading-5 text-muted-foreground">
          推荐 1600x900，高清可用 1920x1080，最低建议 1200x675。
        </span>
      </div>
      <AppButton variant="secondary" size="sm" :disabled="isUploading" @click="$emit('upload')">
        <ImagePlus class="size-4" aria-hidden="true" />
        {{ isUploading ? '上传中...' : '上传' }}
      </AppButton>
    </div>

    <div v-if="cover" class="space-y-3">
      <div
        v-if="inspection?.warnings?.length"
        class="rounded-archive border border-warning/45 bg-warning/10 p-3 text-xs font-semibold leading-5 text-warning"
      >
        <p v-for="warning in inspection.warnings" :key="warning" class="m-0">
          {{ warning }}
        </p>
      </div>

      <div class="rounded-archive border border-border bg-surface-raised p-3">
        <p class="m-0 mb-2 text-xs font-black text-muted-foreground">详情页头图</p>
        <ArticleCover :src="cover" :alt="title || '文章封面'" variant="preview" />
      </div>

      <div class="rounded-archive border border-border bg-surface-raised p-3">
        <p class="m-0 mb-2 text-xs font-black text-muted-foreground">列表卡片</p>
        <div class="grid overflow-hidden rounded-archive border border-border bg-surface md:grid-cols-[8rem_1fr]">
          <ArticleCover :src="cover" :alt="title || '文章封面'" variant="list" fallback-src="/images/go.png" />
          <div class="space-y-2 p-3">
            <div class="h-4 w-3/4 rounded bg-muted" />
            <div class="h-3 w-full rounded bg-muted" />
            <div class="h-3 w-2/3 rounded bg-muted" />
          </div>
        </div>
      </div>

      <div class="rounded-archive border border-border bg-surface-raised p-3">
        <p class="m-0 text-xs font-black text-muted-foreground">分享预览</p>
        <p class="m-0 mt-1 text-xs leading-5 text-muted-foreground">
          当前封面会作为 og:image 和 twitter:image 使用。
        </p>
      </div>

      <div class="flex justify-end">
        <AppButton variant="ghost" size="sm" class="text-danger hover:text-danger" @click="$emit('remove')">
          <Trash2 class="size-4" aria-hidden="true" />
          移除封面
        </AppButton>
      </div>
    </div>

    <div v-else class="rounded-archive border border-dashed border-border bg-surface-raised p-5 text-center">
      <p class="m-0 text-sm font-semibold text-muted-foreground">还没有封面图</p>
    </div>
  </div>
</template>
```

- [ ] **Step 4: Run rendering tests**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverRendering.test.ts
```

Expected: the backend panel contract test passes.

- [ ] **Step 5: Commit backend panel**

Run:

```bash
git add web/frontend/src/admin/editor/AdminArticleCoverPanel.vue web/frontend/src/components/article/articleCoverRendering.test.ts
git commit -m "feat(frontend): add admin cover preview panel"
```

---

### Task 5: Wire Backend Cover Diagnostics

**Files:**
- Modify: `web/frontend/src/views/admin/AdminArticleEditorView.vue`
- Modify: `web/frontend/src/components/article/articleCoverRendering.test.ts`

- [ ] **Step 1: Add source-contract test for editor wiring**

Append this test to `web/frontend/src/components/article/articleCoverRendering.test.ts`:

```typescript
test('admin editor wires cover diagnostics into the cover panel', () => {
  const view = read('views/admin/AdminArticleEditorView.vue')

  expect(view).toContain("import AdminArticleCoverPanel from '@/admin/editor/AdminArticleCoverPanel.vue'")
  expect(view).toContain('inspectArticleCoverDimensions')
  expect(view).toContain('loadArticleCoverDimensions')
  expect(view).toContain('loadArticleCoverFileDimensions')
  expect(view).toContain('coverInspection')
  expect(view).toContain('<AdminArticleCoverPanel')
  expect(view).toContain(':inspection="coverInspection"')
})
```

- [ ] **Step 2: Run rendering tests and verify editor wiring test fails**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverRendering.test.ts
```

Expected: the new test fails because editor wiring is not implemented.

- [ ] **Step 3: Add imports and cover inspection state**

Modify `web/frontend/src/views/admin/AdminArticleEditorView.vue`.

Add imports:

```typescript
import AdminArticleCoverPanel from '@/admin/editor/AdminArticleCoverPanel.vue'
import {
  inspectArticleCoverDimensions,
  loadArticleCoverDimensions,
  loadArticleCoverFileDimensions,
  type ArticleCoverInspection
} from '@/components/article/articleCoverPolicy'
```

Add state near the existing cover refs:

```typescript
const coverInspection = ref<ArticleCoverInspection | null>(null)
let coverInspectionToken = 0
```

- [ ] **Step 4: Add cover inspection helpers**

Add these functions near the existing cover helpers:

```typescript
const inspectCoverFromUrl = async (url: string) => {
  const requestToken = ++coverInspectionToken

  if (!url) {
    coverInspection.value = null
    return
  }

  try {
    const dimensions = await loadArticleCoverDimensions(url)
    if (requestToken !== coverInspectionToken) return
    coverInspection.value = inspectArticleCoverDimensions(dimensions)
  } catch {
    if (requestToken !== coverInspectionToken) return
    coverInspection.value = inspectArticleCoverDimensions(null)
  }
}

const inspectCoverFromFile = async (file: File) => {
  try {
    coverInspection.value = inspectArticleCoverDimensions(await loadArticleCoverFileDimensions(file))
  } catch {
    coverInspection.value = inspectArticleCoverDimensions(null)
  }
}
```

Add this watcher after `coverPreview` is defined:

```typescript
watch(coverPreview, (nextCover) => {
  void inspectCoverFromUrl(nextCover)
})
```

- [ ] **Step 5: Inspect selected cover files during upload**

In `handleCoverInput`, after the `if (!article.value.id) return` guard and before `isCoverUploading.value = true`, add:

```typescript
void inspectCoverFromFile(file!)
```

Keep the existing upload behavior unchanged.

- [ ] **Step 6: Replace inline cover markup with the panel**

In the template, keep the hidden file input and replace the inline cover button/preview block with:

```vue
<div class="space-y-3">
  <input
    ref="coverInput"
    type="file"
    :accept="ADMIN_IMAGE_ACCEPT"
    class="hidden"
    @change="handleCoverInput"
  />
  <AdminArticleCoverPanel
    :cover="coverPreview"
    :title="articleTitle"
    :is-uploading="isCoverUploading"
    :inspection="coverInspection"
    @upload="openCoverPicker"
    @remove="removeCover"
  />
</div>
```

- [ ] **Step 7: Run rendering tests**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverRendering.test.ts
```

Expected: all rendering contract tests pass.

- [ ] **Step 8: Commit backend editor wiring**

Run:

```bash
git add web/frontend/src/views/admin/AdminArticleEditorView.vue web/frontend/src/components/article/articleCoverRendering.test.ts
git commit -m "feat(frontend): surface cover guidance in article editor"
```

---

### Task 6: Preserve SEO Contract And Run Full Verification

**Files:**
- Modify: `web/frontend/src/util/seo.test.ts`

- [ ] **Step 1: Add explicit SEO cover assertions**

Modify the existing `builds article metadata and JSON-LD from rendered content fields` test in `web/frontend/src/util/seo.test.ts` by adding these assertions after the Open Graph image assertion:

```typescript
expect(metadata.twitterCard).toBe('summary_large_image')
expect(metadata.jsonLd).toMatchObject({
  image: 'https://blog.example.com/resources/covers/redis.png'
})
```

Add this new test below it:

```typescript
test('uses summary twitter card when an article has no cover image', () => {
  const metadata = buildArticleSeoMetadata(
    { ...article, cover: '' },
    { siteOrigin: 'https://blog.example.com/' }
  )

  expect(metadata.openGraph?.image).toBeUndefined()
  expect(metadata.twitterCard).toBe('summary')
})
```

- [ ] **Step 2: Run SEO tests**

Run:

```bash
cd web/frontend && bun test src/util/seo.test.ts
```

Expected: all SEO tests pass.

- [ ] **Step 3: Run focused frontend tests**

Run:

```bash
cd web/frontend && bun test src/components/article/articleCoverPolicy.test.ts src/components/article/articleCoverRendering.test.ts src/util/seo.test.ts
```

Expected: all focused tests pass.

- [ ] **Step 4: Run full frontend verification**

Run:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build-only
```

Expected:

- `bun test` passes.
- `bun run type-check` passes.
- `bun run build-only` passes.
- The existing Vite chunk-size warning is acceptable if it appears.

- [ ] **Step 5: Commit SEO and verification test updates**

Run:

```bash
git add web/frontend/src/util/seo.test.ts
git commit -m "test(frontend): cover article image metadata"
```

---

## Final Review Checklist

- [ ] `article.cover` remains the only persisted cover field.
- [ ] Public article detail shows cover above the title only when a cover exists.
- [ ] Public article list no longer uses `object-contain` with padded thumbnails.
- [ ] Backend editor shows recommended cover sizes and non-blocking warnings.
- [ ] Backend editor previews detail, list, and social cover usage.
- [ ] SEO cover metadata remains unchanged except for stronger tests.
- [ ] No backend migration, proto, sqlc, or API changes were introduced.
- [ ] Full frontend test, type-check, and build-only commands pass.
