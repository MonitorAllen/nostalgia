# Editor UX Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden the owner-only article editor by simplifying CKEditor authoring tools and making upload/save feedback consistent.

**Architecture:** Keep changes inside `web/frontend`. Extract pure editor policy and upload policy modules with Bun tests, then wire those policies into CKEditor config, content image upload, cover upload, and save failure handling.

**Tech Stack:** Vue 3, TypeScript, CKEditor 5, Pinia-free pure helpers, Bun test runner, Vite.

---

## File Map

- Create: `web/frontend/src/admin/editor/adminEditorPolicy.ts`
  - Shared toolbar item policy and disallowed toolbar assertions.
- Create: `web/frontend/src/admin/editor/adminEditorPolicy.test.ts`
  - Tests that the toolbar keeps supported tools and excludes risky output tools.
- Modify: `web/frontend/src/admin/editor/adminEditorConfig.ts`
  - Uses the toolbar policy and removes broad/risky CKEditor plugins.
- Create: `web/frontend/src/admin/editor/uploadPolicy.ts`
  - Shared upload validation and error-message helpers.
- Create: `web/frontend/src/admin/editor/uploadPolicy.test.ts`
  - Tests missing file, unsupported MIME, oversized images, accepted JPG/PNG, aborted uploads, and backend error extraction.
- Modify: `web/frontend/src/admin/editor/adminUploadAdapter.ts`
  - Uses `uploadPolicy.ts` before reading content image files.
- Modify: `web/frontend/src/views/admin/AdminArticleEditorView.vue`
  - Uses `uploadPolicy.ts` for cover uploads and save failure toasts.

---

### Task 1: Editor Toolbar Policy

**Files:**
- Create: `web/frontend/src/admin/editor/adminEditorPolicy.test.ts`
- Create: `web/frontend/src/admin/editor/adminEditorPolicy.ts`
- Modify: `web/frontend/src/admin/editor/adminEditorConfig.ts`

- [ ] **Step 1: Write failing policy test**

Create `web/frontend/src/admin/editor/adminEditorPolicy.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import { ADMIN_EDITOR_DISALLOWED_TOOLBAR_ITEMS, ADMIN_EDITOR_TOOLBAR_ITEMS } from './adminEditorPolicy'

describe('ADMIN_EDITOR_TOOLBAR_ITEMS', () => {
  test('keeps core blog authoring tools', () => {
    expect(ADMIN_EDITOR_TOOLBAR_ITEMS).toContain('heading')
    expect(ADMIN_EDITOR_TOOLBAR_ITEMS).toContain('bold')
    expect(ADMIN_EDITOR_TOOLBAR_ITEMS).toContain('italic')
    expect(ADMIN_EDITOR_TOOLBAR_ITEMS).toContain('insertImage')
    expect(ADMIN_EDITOR_TOOLBAR_ITEMS).toContain('insertTable')
    expect(ADMIN_EDITOR_TOOLBAR_ITEMS).toContain('blockQuote')
    expect(ADMIN_EDITOR_TOOLBAR_ITEMS).toContain('codeBlock')
  })

  test('does not expose tools that produce unsupported reader output', () => {
    ADMIN_EDITOR_DISALLOWED_TOOLBAR_ITEMS.forEach((item) => {
      expect(ADMIN_EDITOR_TOOLBAR_ITEMS).not.toContain(item)
    })
  })
})
```

- [ ] **Step 2: Run red check**

Run:

```bash
cd web/frontend && bun test src/admin/editor/adminEditorPolicy.test.ts
```

Expected: FAIL because `adminEditorPolicy.ts` does not exist.

- [ ] **Step 3: Implement toolbar policy**

Create `web/frontend/src/admin/editor/adminEditorPolicy.ts`:

```ts
export const ADMIN_EDITOR_TOOLBAR_ITEMS = [
  'undo',
  'redo',
  '|',
  'heading',
  '|',
  'bold',
  'italic',
  'underline',
  'strikethrough',
  'removeFormat',
  '|',
  'bulletedList',
  'numberedList',
  'todoList',
  '|',
  'link',
  'insertImage',
  'insertTable',
  'blockQuote',
  'codeBlock',
  'horizontalLine',
  '|',
  'alignment'
] as const

export const ADMIN_EDITOR_DISALLOWED_TOOLBAR_ITEMS = [
  'fontFamily',
  'fontSize',
  'fontColor',
  'fontBackgroundColor',
  'htmlEmbed',
  'sourceEditing',
  'mediaEmbed',
  'showBlocks',
  'subscript',
  'superscript'
] as const
```

- [ ] **Step 4: Wire toolbar and remove broad plugins**

In `adminEditorConfig.ts`, import `ADMIN_EDITOR_TOOLBAR_ITEMS`, set:

```ts
toolbar: {
  items: [...ADMIN_EDITOR_TOOLBAR_ITEMS],
  shouldNotGroupWhenFull: false
}
```

Remove imports and plugin entries for:

```text
Autosave, FontBackgroundColor, FontColor, FontFamily, FontSize, FullPage,
GeneralHtmlSupport, Highlight, HtmlComment, HtmlEmbed, Indent, IndentBlock,
Markdown, MediaEmbed, PasteFromMarkdownExperimental, ShowBlocks,
SimpleUploadAdapter, SourceEditing, SpecialCharacters*, Subscript, Superscript
```

Remove `fontFamily` and `fontSize` config blocks.

- [ ] **Step 5: Verify**

Run:

```bash
cd web/frontend && bun test src/admin/editor/adminEditorPolicy.test.ts
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: all commands exit 0.

- [ ] **Step 6: Commit**

```bash
git add web/frontend/src/admin/editor/adminEditorPolicy.ts web/frontend/src/admin/editor/adminEditorPolicy.test.ts web/frontend/src/admin/editor/adminEditorConfig.ts
git commit -m "refactor(frontend): simplify admin editor tools"
```

---

### Task 2: Upload And Save Feedback Policy

**Files:**
- Create: `web/frontend/src/admin/editor/uploadPolicy.test.ts`
- Create: `web/frontend/src/admin/editor/uploadPolicy.ts`
- Modify: `web/frontend/src/admin/editor/adminUploadAdapter.ts`
- Modify: `web/frontend/src/views/admin/AdminArticleEditorView.vue`

- [ ] **Step 1: Write failing upload policy tests**

Create `web/frontend/src/admin/editor/uploadPolicy.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import { getAdminUploadErrorMessage, validateAdminImageFile } from './uploadPolicy'

describe('validateAdminImageFile', () => {
  test('requires a file', () => {
    expect(validateAdminImageFile()).toBe('请选择要上传的图片')
  })

  test('rejects unsupported image types', () => {
    expect(validateAdminImageFile({ type: 'image/webp', size: 1024 })).toBe('仅支持 JPG 或 PNG 图片')
  })

  test('rejects oversized images', () => {
    expect(validateAdminImageFile({ type: 'image/png', size: 5 * 1024 * 1024 + 1 })).toBe('图片不能超过 5 MB')
  })

  test('accepts jpg and png images within the size limit', () => {
    expect(validateAdminImageFile({ type: 'image/jpeg', size: 1024 })).toBe('')
    expect(validateAdminImageFile({ type: 'image/png', size: 1024 })).toBe('')
  })
})

describe('getAdminUploadErrorMessage', () => {
  test('uses backend error text when available', () => {
    expect(getAdminUploadErrorMessage({ response: { data: { error: '不支持的文件类型' } } })).toBe('不支持的文件类型')
  })

  test('uses abort message for canceled uploads', () => {
    expect(getAdminUploadErrorMessage({ name: 'AbortError' })).toBe('上传已取消')
  })
})
```

- [ ] **Step 2: Run red check**

Run:

```bash
cd web/frontend && bun test src/admin/editor/uploadPolicy.test.ts
```

Expected: FAIL because `uploadPolicy.ts` does not exist.

- [ ] **Step 3: Implement upload policy**

Create `web/frontend/src/admin/editor/uploadPolicy.ts` with:

```ts
export const ADMIN_IMAGE_MAX_BYTES = 5 * 1024 * 1024
export const ADMIN_IMAGE_ALLOWED_TYPES = ['image/jpeg', 'image/png'] as const

export type AdminImageFileLike = Pick<File, 'size' | 'type'>

export function validateAdminImageFile(file?: AdminImageFileLike | null) {
  if (!file) return '请选择要上传的图片'
  if (!ADMIN_IMAGE_ALLOWED_TYPES.includes(file.type as (typeof ADMIN_IMAGE_ALLOWED_TYPES)[number])) {
    return '仅支持 JPG 或 PNG 图片'
  }
  if (file.size > ADMIN_IMAGE_MAX_BYTES) return '图片不能超过 5 MB'
  return ''
}

export function getAdminUploadErrorMessage(error: unknown, fallback = '请稍后再试') {
  if (typeof error === 'object' && error) {
    if ('name' in error && error.name === 'AbortError') return '上传已取消'

    const response = (error as { response?: { data?: { error?: string; message?: string } | string } }).response
    const data = response?.data
    if (typeof data === 'string' && data) return data
    if (data?.error) return data.error
    if (data?.message) return data.message

    if ('message' in error && typeof error.message === 'string' && error.message) {
      return error.message
    }
  }

  return fallback
}
```

- [ ] **Step 4: Wire upload policy**

In `adminUploadAdapter.ts`, validate before reading the file and wrap upload failures through `getAdminUploadErrorMessage`.

In `AdminArticleEditorView.vue`, replace local `validateImageFile` with `validateAdminImageFile`, and use `getAdminUploadErrorMessage` for cover upload and article save failures.

- [ ] **Step 5: Verify**

Run:

```bash
cd web/frontend && bun test src/admin/editor/uploadPolicy.test.ts
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: all commands exit 0.

- [ ] **Step 6: Commit**

```bash
git add web/frontend/src/admin/editor/uploadPolicy.ts web/frontend/src/admin/editor/uploadPolicy.test.ts web/frontend/src/admin/editor/adminUploadAdapter.ts web/frontend/src/views/admin/AdminArticleEditorView.vue
git commit -m "fix(frontend): harden editor upload feedback"
```

---

### Task 3: Final Verification

**Files:**
- No planned source edits.

- [ ] **Step 1: Run frontend verification**

Run:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: all commands exit 0.

- [ ] **Step 2: Review branch state**

Run:

```bash
git status --short --branch
git log --oneline --decorate --max-count=8
```

Expected: branch is `feature/editor-ux-hardening`; working tree is clean after commits.
