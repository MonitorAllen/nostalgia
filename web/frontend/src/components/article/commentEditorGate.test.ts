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
