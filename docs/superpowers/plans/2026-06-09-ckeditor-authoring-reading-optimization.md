# CKEditor Authoring Reading Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make CKEditor authoring, public article reading, and comments feel consistent, safer, and more comfortable to use.

**Architecture:** Keep the backend admin editor and frontend reader as separate Vue applications, but share the same content styling vocabulary through mirrored content CSS files. Fix upload cleanup in the Go utility layer with focused tests, then layer editor UX improvements around existing CKEditor and Element Plus patterns.

**Tech Stack:** Go, Vue 3, TypeScript, CKEditor 5, Element Plus, Tailwind CSS, Bun, Vite.

---

### Task 1: Resource Filename Extraction

**Files:**
- Create: `util/file_test.go`
- Modify: `util/file.go`

- [ ] **Step 1: Write a failing test for article resource URLs**

```go
func TestExtractFileNamesFindsArticleResourceNames(t *testing.T) {
	content := `
		<img src="/resources/articles/6d5f/relative-image.png">
		<img src="http://localhost:8080/resources/articles/6d5f/absolute-image.jpg?t=123">
		<a href="https://example.com/resources/articles/6d5f/document.pdf#page=1">file</a>
		<img src="/resources/articles/6d5f/relative-image.png">
		<a href="https://example.com/not-an-upload">external link</a>
	`

	got := ExtractFileNames(content)
	want := []string{"relative-image.png", "absolute-image.jpg", "document.pdf"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractFileNames() = %#v, want %#v", got, want)
	}
}
```

- [ ] **Step 2: Run the focused test and confirm it fails**

Run: `go test ./util -run TestExtractFileNamesFindsArticleResourceNames -count=1`

Expected: FAIL because the old implementation misses relative `/resources/...` URLs and keeps query or hash suffixes.

- [ ] **Step 3: Implement resource-only extraction**

Parse HTML attributes such as `src`, `href`, `poster`, and `data-src`, normalize URL paths, accept only paths containing `/resources/`, unescape the basename, and de-duplicate in content order.

- [ ] **Step 4: Verify the focused test passes**

Run: `go test ./util -run TestExtractFileNamesFindsArticleResourceNames -count=1`

Expected: PASS.

### Task 2: Shared Content Typography

**Files:**
- Create: `web/frontend/src/assets/content.css`
- Create: `web/backend/src/styles/content.scss`
- Modify: `web/frontend/src/assets/main.css`
- Modify: `web/backend/src/views/article/editor/index.vue`

- [ ] **Step 1: Extract frontend `.reading-prose` rules**

Move article-content typography, headings, blockquote, inline code, pre, image, figure, caption, list, table, link, and horizontal rule styles from `main.css` into `content.css`.

- [ ] **Step 2: Import frontend content CSS**

Import `./content.css` from `web/frontend/src/assets/main.css` after Tailwind directives.

- [ ] **Step 3: Add matching backend content SCSS**

Create `.nostalgia-content` and `.ck-editor__editable.nostalgia-content` styles in `web/backend/src/styles/content.scss` with the same rhythm and surface treatment, adapted to Element Plus tokens.

- [ ] **Step 4: Attach the backend content class to CKEditor**

Use CKEditor `onReady` to add `nostalgia-content` to the editing root and import the backend content stylesheet from the editor page.

### Task 3: Backend Editor UX

**Files:**
- Modify: `web/backend/src/config/editorConfig.ts`
- Modify: `web/backend/src/views/article/editor/index.vue`
- Modify: `web/backend/src/utils/uploadAdapter.ts`

- [ ] **Step 1: Simplify the toolbar**

Keep writing tools visible: undo, redo, heading, bold, italic, underline, lists, quote, link, image, table, code block, alignment, remove format, and source editing.

- [ ] **Step 2: Add save-state feedback**

Track `isDirty`, `saveStatus`, `lastSaveTime`, and `hasDraft`. Show saved, unsaved, saving, and failed states in the footer and disable double saves.

- [ ] **Step 3: Add keyboard and unload protection**

Bind Ctrl/Cmd+S to `saveArticle`, and warn before closing when there are unsaved changes.

- [ ] **Step 4: Improve upload validation and errors**

Reject unsupported images before base64 conversion, set CKEditor loader upload totals, and surface friendly Element Plus messages for content and cover upload failures.

### Task 4: Frontend Comment Editor and Highlighting

**Files:**
- Modify: `web/frontend/src/views/article/ArticleView.vue`
- Modify: `web/frontend/src/components/article/CommentItem.vue`

- [ ] **Step 1: Align Prism language imports**

Import Prism languages that match backend CKEditor code block options: Go, Python, JavaScript, TypeScript, Java, C, C++, SQL, JSON, Bash, HTML, and CSS.

- [ ] **Step 2: Improve comment editor behavior**

Add placeholder text, Ctrl/Cmd+Enter submit, and HTML-aware empty-content validation.

- [ ] **Step 3: Apply comment content sizing**

Render comments with the same content style class plus a compact modifier so comments stay readable without inheriting full article scale.

### Task 5: Verification and Commits

**Files:**
- Modify only files touched by the tasks above.

- [ ] **Step 1: Run Go verification**

Run: `go test ./util -count=1`

- [ ] **Step 2: Run backend admin verification**

Run: `cd web/backend && npm run type:check`

- [ ] **Step 3: Run frontend verification**

Run: `cd web/frontend && bun run type-check && bun run build`

- [ ] **Step 4: Split commits**

Create logical Conventional Commits for backend cleanup, shared content/editor UX, and frontend comment/reader polish.
