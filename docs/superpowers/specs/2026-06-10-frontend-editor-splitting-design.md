# Frontend Editor Splitting Design

## Context

The unified Vue frontend now serves both public reading pages and the owner-only `/admin` area. CKEditor is required for article authoring and authenticated commenting, but the public article page currently imports CKEditor directly. That makes reader traffic pay for editor code even when a visitor only reads an article or is not logged in.

This branch reduces the public reader path by moving the comment editor behind a lazy boundary. Admin authoring stays unchanged except for removing any unnecessary global CKEditor registration from the app entry when local component imports already cover editor usage.

## Goals

- Remove CKEditor and CKEditor CSS imports from `ArticleView.vue`.
- Load the public comment editor only when a logged-in user explicitly activates the comment box or starts a reply.
- Keep guests on a lightweight login/register prompt without loading the comment editor.
- Preserve current comment behavior: empty-content validation, reply targeting, cancel reply, submit button state, `Ctrl/Cmd + Enter`, Chinese placeholder, code/codeBlock tools, and shared code block language list.
- Keep admin editor functionality intact.
- Make the bundle split visible in Vite output with CKEditor code isolated away from the public article route module.

## Non-Goals

- No backend API changes.
- No database or migration changes.
- No CKEditor version upgrade.
- No replacement of CKEditor.
- No visual redesign of comments beyond the small activation state needed for lazy loading.
- No removal of the legacy `EditorView.vue` file unless it is separately proven unused and approved.

## Design

### Lazy Comment Editor

Create `web/frontend/src/components/article/CommentEditor.vue` as the only public-comment component that imports:

- `@ckeditor/ckeditor5-vue`
- `ckeditor5`
- `ckeditor5/translations/zh-cn.js`
- `ckeditor5/ckeditor5.css`
- `ckeditor5/ckeditor5-content.css`

`ArticleView.vue` imports this component through `defineAsyncComponent`. Because the component is rendered only behind an authenticated-and-activated gate, guests and passive readers should not fetch the CKEditor chunk.

### Activation Gate

Add a small pure helper in `web/frontend/src/components/article/commentEditorGate.ts`:

- authenticated + active: render the lazy editor.
- authenticated + inactive: render a lightweight "write comment" action.
- guest: render the existing login/register prompt.

`ArticleView.vue` owns the activation state. Clicking the write-comment action activates the editor. Clicking reply while logged in also activates it, preserves reply metadata, and scrolls to the editor after the DOM updates.

### Comment Flow

`CommentEditor.vue` receives `modelValue` and `disabled`, emits `update:modelValue`, and emits `submit` when the user presses `Ctrl/Cmd + Enter`. `ArticleView.vue` keeps comment submission, reply resolution, store calls, toast messages, and comment-list mutation in the parent so API behavior remains unchanged.

After a successful submission, `ArticleView.vue` clears `editorData` and resets reply state as it does today. The editor may remain active so the logged-in user can continue writing without paying the lazy-load cost again.

### App Entry

Remove the global `CkeditorPlugin` registration from `web/frontend/src/main.ts` if the build and type-check show local component imports are sufficient. Admin editor already imports `Ckeditor` locally, and the public lazy editor will also import it locally.

### Verification

Minimum verification:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

The build output should show a separate async editor-related chunk instead of keeping the public article module tied to direct CKEditor imports. A remaining large-chunk warning is acceptable if the main reader path is reduced and editor code is split.
