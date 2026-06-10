# Editor Reading Parity Design

## Context

The admin frontend has been migrated into `web/frontend`. The old CKEditor optimization plan still mentions `web/backend`, but the current product direction is one Vue app with public reading pages and owner-only `/admin` authoring.

The current implementation already uses `reading-prose` for public article content, compact comment rendering, and the admin CKEditor editable root. The remaining problem is polish and consistency: wide editor lines can become too long, comment authoring does not use the compact prose class, and CKEditor-generated image/table figure classes need explicit public reader styling.

## Goals

- Make public article rendering, comment rendering, comment authoring, and admin authoring share the same prose vocabulary.
- Keep readable line length on wide admin screens without shrinking code blocks, tables, or images.
- Share CKEditor code block language options between admin articles and public comments.
- Add explicit styles for CKEditor figures, image alignment classes, captions, and table wrappers.
- Preserve the current glass archive theme while keeping article body surfaces solid and comfortable.

## Non-Goals

- No backend API or database changes.
- No CKEditor package replacement.
- No redesign of the whole admin article editor layout.
- No visual overhaul of navigation, article cards, or theme tokens.
- No work in removed legacy `web/backend`.

## Content Styling Model

`web/frontend/src/assets/content.css` remains the source of truth for authored content. The same classes should be used in all content contexts:

- public article body: `reading-prose ck-content`
- rendered comments: `reading-prose reading-prose--compact ck-content`
- comment editor editable root: `reading-prose reading-prose--compact comment-editor-content`
- admin article editor editable root: `reading-prose ck-content` inside `admin-editor-content`

Direct textual children should use a prose measure around 72ch. Code blocks, tables, and images may use the full available width because technical content often needs horizontal space.

## CKEditor Output Support

The reader should intentionally style CKEditor output rather than relying only on upstream default CSS:

- `figure.image`
- `figure.table`
- `figcaption`
- `image-style-align-left`
- `image-style-align-right`
- `image-style-align-center`
- `image-style-side`

On mobile, floated image styles should collapse to full-width blocks to prevent cramped text columns.

## Code Block Languages

Admin article editing and public comment editing should import the same code block language list from a shared frontend module. Prism imports remain in `ArticleView.vue`, but the selectable CKEditor languages should not be duplicated.

The shared list is:

```text
plaintext, go, python, javascript, typescript, java, c, cpp, sql, json, bash, html, css
```

## Verification

Minimum verification:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

The existing Vite large chunk warning is acceptable for this branch because CKEditor is already part of the app and this work does not increase the editor dependency surface.
