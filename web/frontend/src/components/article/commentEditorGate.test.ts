import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

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

describe('article comment composer contract', () => {
  test('shows the submit actions only after the comment editor is active', () => {
    const source = readFileSync(
      resolve(import.meta.dir, '../../views/article/ArticleView.vue'),
      'utf8'
    )

    expect(source).toContain('<CommentEditor\n          v-if="canRenderCommentEditor"')
    expect(source).toContain('<div v-if="canRenderCommentEditor" class="mt-3')
    expect(source).toContain('<AppButton size="sm" class="w-max" @click="activateCommentEditor">写评论</AppButton>')
    expect(source).toContain("replyCommentId === 0 ? '发表评论' : '回复评论'")
  })
})
