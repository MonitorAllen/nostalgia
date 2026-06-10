# Editor Reading Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align CKEditor authoring, public article reading, and comment rendering around one shared content style system.

**Architecture:** Keep all work inside the unified `web/frontend` app. Extract CKEditor code block language options into a shared module, then tighten `content.css` so public reader output and admin editor output handle prose measure, figures, tables, images, and compact comment editing consistently.

**Tech Stack:** Vue 3, TypeScript, CKEditor 5, Tailwind CSS tokens, Bun test runner, Vite.

---

## File Map

- Create: `web/frontend/src/editor/contentLanguages.ts`
  - Shared CKEditor code block language option list.
- Create: `web/frontend/src/editor/contentLanguages.test.ts`
  - Tests language coverage and duplicate prevention.
- Modify: `web/frontend/src/admin/editor/adminEditorConfig.ts`
  - Imports the shared language list.
- Modify: `web/frontend/src/views/article/ArticleView.vue`
  - Imports shared language list and applies compact prose class to the comment editor editable root.
- Modify: `web/frontend/src/assets/content.css`
  - Adds prose measure, CKEditor figure/table/image alignment support, and editor toolbar/content polish.

---

### Task 1: Shared Code Block Language Options

**Files:**
- Create: `web/frontend/src/editor/contentLanguages.test.ts`
- Create: `web/frontend/src/editor/contentLanguages.ts`
- Modify: `web/frontend/src/admin/editor/adminEditorConfig.ts`
- Modify: `web/frontend/src/views/article/ArticleView.vue`

- [ ] **Step 1: Write failing Bun test**

Create `web/frontend/src/editor/contentLanguages.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import { CODE_BLOCK_LANGUAGES } from './contentLanguages'

describe('CODE_BLOCK_LANGUAGES', () => {
  test('keeps article and comment editors aligned with expected languages', () => {
    const languages = CODE_BLOCK_LANGUAGES.map((item) => item.language)

    expect(languages).toEqual([
      'plaintext',
      'go',
      'python',
      'javascript',
      'typescript',
      'java',
      'c',
      'cpp',
      'sql',
      'json',
      'bash',
      'html',
      'css'
    ])
    expect(new Set(languages).size).toBe(languages.length)
  })
})
```

- [ ] **Step 2: Run red check**

Run:

```bash
cd web/frontend && bun test src/editor/contentLanguages.test.ts
```

Expected: FAIL because `contentLanguages.ts` does not exist.

- [ ] **Step 3: Implement shared module**

Create `web/frontend/src/editor/contentLanguages.ts`:

```ts
export const CODE_BLOCK_LANGUAGES = [
  { language: 'plaintext', label: 'Plain text' },
  { language: 'go', label: 'Golang' },
  { language: 'python', label: 'Python' },
  { language: 'javascript', label: 'JavaScript' },
  { language: 'typescript', label: 'TypeScript' },
  { language: 'java', label: 'Java' },
  { language: 'c', label: 'C' },
  { language: 'cpp', label: 'C++' },
  { language: 'sql', label: 'SQL' },
  { language: 'json', label: 'JSON' },
  { language: 'bash', label: 'Bash' },
  { language: 'html', label: 'HTML' },
  { language: 'css', label: 'CSS' }
] as const
```

- [ ] **Step 4: Use the shared list**

Import `CODE_BLOCK_LANGUAGES` in:

```ts
web/frontend/src/admin/editor/adminEditorConfig.ts
web/frontend/src/views/article/ArticleView.vue
```

Replace duplicated inline `codeBlock.languages` arrays with:

```ts
languages: [...CODE_BLOCK_LANGUAGES]
```

- [ ] **Step 5: Verify**

Run:

```bash
cd web/frontend && bun test src/editor/contentLanguages.test.ts
cd web/frontend && bun run type-check
```

Expected: both commands exit 0.

- [ ] **Step 6: Commit**

```bash
git add web/frontend/src/editor/contentLanguages.ts web/frontend/src/editor/contentLanguages.test.ts web/frontend/src/admin/editor/adminEditorConfig.ts web/frontend/src/views/article/ArticleView.vue
git commit -m "refactor(frontend): share editor code block languages"
```

---

### Task 2: Content CSS Parity

**Files:**
- Modify: `web/frontend/src/assets/content.css`
- Modify: `web/frontend/src/views/article/ArticleView.vue`

- [ ] **Step 1: Add prose measure and CKEditor figure support**

Update `content.css` so direct text children use `--prose-measure: 72ch`, while `pre`, `figure`, and `table` keep full-width behavior. Add explicit support for CKEditor image alignment and table figures.

- [ ] **Step 2: Apply compact prose to the comment editor**

In `ArticleView.vue`, update `onEditorReady`:

```ts
editorInstance.ui.view.editable.element?.classList.add(
  'reading-prose',
  'reading-prose--compact',
  'comment-editor-content'
)
```

- [ ] **Step 3: Verify frontend**

Run:

```bash
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: both commands exit 0.

- [ ] **Step 4: Commit**

```bash
git add web/frontend/src/assets/content.css web/frontend/src/views/article/ArticleView.vue
git commit -m "feat(frontend): align editor and reader content styles"
```

---

### Task 3: Final Verification

**Files:**
- No planned source edits.

- [ ] **Step 1: Run frontend tests and build**

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

Expected: branch is `feature/editor-reading-parity`; working tree is clean after commits.
