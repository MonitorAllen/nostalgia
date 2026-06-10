# Editor UX Hardening Design

## Context

The editor and reader now share content styling, and `/admin` lives inside the unified `web/frontend` app. The next problem is workflow quality: the admin editor still exposes a broad CKEditor plugin set, upload validation is duplicated between cover and content upload, and save/upload failures can be clearer.

This branch hardens the existing owner-only editor without changing backend APIs, database schema, or the public article layout.

## Goals

- Reduce the admin CKEditor toolbar and plugin surface to blog-appropriate authoring tools.
- Remove editor capabilities that are likely to produce HTML that the public sanitizer does not preserve or that conflicts with the public reading style.
- Share one upload validation policy for cover uploads and CKEditor content image uploads.
- Show friendly, consistent messages for unsupported image formats, oversized files, canceled uploads, backend upload failures, and save failures.
- Keep current draft recovery, keyboard save shortcut, and unsaved-leave guard behavior.

## Non-Goals

- No lazy loading or bundle splitting work in this branch.
- No backend upload API changes.
- No article schema changes.
- No replacing CKEditor.
- No redesign of the whole admin editor shell.

## Editor Policy

Keep tools that produce stable blog content:

- undo, redo
- heading
- bold, italic, underline, strikethrough, remove format
- bullet list, numbered list, todo list
- link
- image upload and insert
- table insert and table editing
- blockquote
- code block
- horizontal line
- alignment

Remove or stop exposing tools that encourage unsupported or inconsistent output:

- font family, font size, font color, font background
- full page editing
- raw HTML embed and HTML comments
- source editing
- markdown mode and paste-from-markdown experimental mode
- media embed
- show blocks
- special character packs
- subscript and superscript
- simple upload adapter, because the app uses its own admin upload adapter

The editor should remain capable of loading existing saved HTML, but the authoring surface should guide new content into the supported public reader vocabulary.

## Upload Policy

The frontend should validate image uploads before reading the file:

- Missing file: `请选择要上传的图片`
- Unsupported type: `仅支持 JPG 或 PNG 图片`
- Oversized file: `图片不能超过 5 MB`

The policy applies to both cover image uploads and CKEditor content image uploads.

Backend errors should still pass through when available. If an upload is aborted, the message should be `上传已取消`.

## Save Feedback

When article saving fails, keep the existing draft cache and `saveStatus = 'error'`, but also show a toast near the action flow:

- summary: `保存失败`
- detail: backend `error` or `message` when available, otherwise `修改已保存在本地草稿，请稍后重试`

## Verification

Minimum verification:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

The existing Vite large chunk warning remains acceptable here. Bundle splitting is the next separate branch.
