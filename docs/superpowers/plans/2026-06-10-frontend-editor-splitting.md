# Frontend Editor Splitting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lazy-load the public comment CKEditor so visitors and passive readers do not load editor code on the article reading path.

**Architecture:** Keep comment API behavior in `ArticleView.vue`, extract CKEditor rendering into a lazy `CommentEditor.vue`, and guard rendering with a small tested pure helper. Remove global CKEditor plugin registration only after local component imports cover all active editor usage.

**Tech Stack:** Vue 3, TypeScript, CKEditor 5, Bun test runner, Vite.

---

## File Map

- Create: `web/frontend/src/components/article/commentEditorGate.ts`
  - Pure render-gating helper for public comment editor loading.
- Create: `web/frontend/src/components/article/commentEditorGate.test.ts`
  - Bun tests for guest, authenticated inactive, and authenticated active states.
- Create: `web/frontend/src/components/article/CommentEditor.vue`
  - Lazy public comment CKEditor wrapper with local CKEditor imports and submit shortcut.
- Modify: `web/frontend/src/views/article/ArticleView.vue`
  - Remove static CKEditor imports/config, add async component gate, and keep comment submission behavior in the parent.
- Modify: `web/frontend/src/main.ts`
  - Remove global `CkeditorPlugin` registration after verifying local imports work.

---

### Task 1: Add Editor Render Gate

**Files:**
- Create: `web/frontend/src/components/article/commentEditorGate.test.ts`
- Create: `web/frontend/src/components/article/commentEditorGate.ts`

- [ ] **Step 1: Write the failing gate test**

Create `web/frontend/src/components/article/commentEditorGate.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import { shouldRenderCommentEditor } from './commentEditorGate'

describe('shouldRenderCommentEditor', () => {
  test('keeps the editor unloaded for guests', () => {
    expect(shouldRenderCommentEditor({ isAuthenticated: false, isActivated: false })).toBe(false)
    expect(shouldRenderCommentEditor({ isAuthenticated: false, isActivated: true })).toBe(false)
  })

  test('keeps the editor unloaded for authenticated passive readers', () => {
    expect(shouldRenderCommentEditor({ isAuthenticated: true, isActivated: false })).toBe(false)
  })

  test('renders the editor only after an authenticated user activates it', () => {
    expect(shouldRenderCommentEditor({ isAuthenticated: true, isActivated: true })).toBe(true)
  })
})
```

- [ ] **Step 2: Run red check**

Run:

```bash
cd web/frontend && bun test src/components/article/commentEditorGate.test.ts
```

Expected: FAIL because `commentEditorGate.ts` does not exist.

- [ ] **Step 3: Implement the gate helper**

Create `web/frontend/src/components/article/commentEditorGate.ts`:

```ts
export interface CommentEditorGateState {
  isAuthenticated: boolean
  isActivated: boolean
}

export const shouldRenderCommentEditor = ({
  isAuthenticated,
  isActivated
}: CommentEditorGateState) => isAuthenticated && isActivated
```

- [ ] **Step 4: Run green check**

Run:

```bash
cd web/frontend && bun test src/components/article/commentEditorGate.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/article/commentEditorGate.ts web/frontend/src/components/article/commentEditorGate.test.ts
git commit -m "test(frontend): cover comment editor lazy gate"
```

### Task 2: Extract Lazy Comment Editor

**Files:**
- Create: `web/frontend/src/components/article/CommentEditor.vue`
- Modify: `web/frontend/src/views/article/ArticleView.vue`

- [ ] **Step 1: Create the lazy editor component**

Create `web/frontend/src/components/article/CommentEditor.vue`:

```vue
<script setup lang="ts">
import { computed, onMounted, ref, type Ref } from 'vue'
import { Ckeditor } from '@ckeditor/ckeditor5-vue'
import { ClassicEditor, Code, CodeBlock, type EditorConfig, Essentials, Paragraph } from 'ckeditor5'
import translations from 'ckeditor5/translations/zh-cn.js'
import 'ckeditor5/ckeditor5.css'
import 'ckeditor5/ckeditor5-content.css'

import { CODE_BLOCK_LANGUAGES } from '@/editor/contentLanguages'

const props = withDefaults(
  defineProps<{
    modelValue: string
    disabled?: boolean
  }>(),
  {
    disabled: false
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  submit: []
}>()

const editorData = computed({
  get: () => props.modelValue,
  set: (value: string) => emit('update:modelValue', value)
})

const isLayoutReady = ref(false)

const config: Ref<EditorConfig> = ref({
  toolbar: {
    items: ['undo', 'redo', '|', 'code', 'codeBlock'],
    shouldNotGroupWhenFull: true
  },
  plugins: [Code, CodeBlock, Essentials, Paragraph],
  placeholder: '写下评论，Ctrl/⌘ + Enter 提交',
  codeBlock: {
    languages: [...CODE_BLOCK_LANGUAGES]
  },
  language: 'zh-cn',
  translations: [translations]
})

const onEditorReady = (editorInstance: ClassicEditor) => {
  editorInstance.ui.view.editable.element?.classList.add(
    'reading-prose',
    'reading-prose--compact',
    'comment-editor-content'
  )
  editorInstance.editing.view.document.on('keydown', (event: unknown, data: any) => {
    const domEvent = data.domEvent as KeyboardEvent
    if ((domEvent.ctrlKey || domEvent.metaKey) && domEvent.key === 'Enter') {
      data.preventDefault()
      ;(event as { stop: () => void }).stop()
      emit('submit')
    }
  })
}

onMounted(() => {
  isLayoutReady.value = true
})
</script>

<template>
  <div id="comment-editor" class="overflow-hidden rounded-archive border border-border">
    <Ckeditor
      v-if="isLayoutReady"
      v-model="editorData"
      :editor="ClassicEditor"
      :config="config"
      :disabled="disabled"
      @ready="onEditorReady"
    />
  </div>
</template>

<style scoped>
:deep(.ck-editor__editable_inline) {
  min-height: 8rem;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-foreground));
}

:deep(.ck-toolbar) {
  background: rgb(var(--color-surface-raised)) !important;
  border-color: rgb(var(--color-border)) !important;
}

:deep(.ck-content) {
  background: rgb(var(--color-surface)) !important;
}
</style>
```

- [ ] **Step 2: Wire the async component in `ArticleView.vue`**

In `ArticleView.vue`, remove direct CKEditor imports and replace them with:

```ts
import { computed, defineAsyncComponent, nextTick, onMounted, onUnmounted, onUpdated, provide, ref } from 'vue'
import { shouldRenderCommentEditor } from '@/components/article/commentEditorGate'

const CommentEditor = defineAsyncComponent(() => import('@/components/article/CommentEditor.vue'))
const isCommentEditorActive = ref(false)
const canRenderCommentEditor = computed(() =>
  shouldRenderCommentEditor({
    isAuthenticated: Boolean(userStore.userInfo),
    isActivated: isCommentEditorActive.value
  })
)
```

Replace the old `onEditorReady`, editor config, editor ref, and manual editor destroy logic with parent-owned activation:

```ts
const scrollToCommentEditor = async () => {
  await nextTick()
  document.getElementById('comment-editor')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

const activateCommentEditor = async () => {
  if (!userStore.userInfo) {
    toast.add({
      severity: 'info',
      summary: '需要登录',
      detail: '登录后才能使用评论功能',
      life: 2500
    })
    return false
  }

  isCommentEditorActive.value = true
  await scrollToCommentEditor()
  return true
}
```

Update logged-in comment template to render a lightweight activation state until `canRenderCommentEditor` is true. Once true, render:

```vue
<CommentEditor
  v-model="editorData"
  :disabled="isSubmittingComment"
  @submit="createComment(0, article.owner)"
/>
```

- [ ] **Step 3: Preserve reply behavior**

Update `replyComment` so the unauthenticated path keeps the existing login toast, cancel reply still clears state, and logged-in reply activates the editor before scrolling:

```ts
isCommentEditorActive.value = true
replyCommentId.value = commentId
replyUserName.value = `@${toUserName}`
replyUserId.value = toUserId
replyCommentParentId.value = parentId === 0 ? commentId : parentId
void scrollToCommentEditor()
```

- [ ] **Step 4: Run type check**

Run:

```bash
cd web/frontend && bun run type-check
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/frontend/src/components/article/CommentEditor.vue web/frontend/src/views/article/ArticleView.vue
git commit -m "perf(frontend): lazy load public comment editor"
```

### Task 3: Remove Global CKEditor Registration

**Files:**
- Modify: `web/frontend/src/main.ts`

- [ ] **Step 1: Remove global plugin import and registration**

In `web/frontend/src/main.ts`, delete:

```ts
import { CkeditorPlugin } from '@ckeditor/ckeditor5-vue'
app.use(CkeditorPlugin)
```

- [ ] **Step 2: Verify active editors still compile**

Run:

```bash
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: PASS. The admin editor and lazy comment editor both import `Ckeditor` locally.

- [ ] **Step 3: Commit**

```bash
git add web/frontend/src/main.ts
git commit -m "perf(frontend): remove global ckeditor registration"
```

### Task 4: Final Verification and PR

**Files:**
- Verify changed files only; no new production files expected.

- [ ] **Step 1: Run full frontend checks**

Run:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: all commands exit 0.

- [ ] **Step 2: Inspect bundle output**

Check the Vite output from `bun run build`. Expected: public article code no longer directly imports CKEditor, and the build output contains a separate async component/editor chunk.

- [ ] **Step 3: Push and open PR**

Run:

```bash
git status --short --branch
git push -u origin perf/frontend-editor-splitting
gh pr create --base master --head perf/frontend-editor-splitting --title "perf(frontend): lazy load public comment editor" --body-file /tmp/nostalgia-frontend-editor-splitting-pr.md
```
