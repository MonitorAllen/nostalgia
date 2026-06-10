import { describe, expect, test } from 'bun:test'
import {
  ADMIN_EDITOR_DISALLOWED_TOOLBAR_ITEMS,
  ADMIN_EDITOR_TOOLBAR_ITEMS,
} from './adminEditorPolicy'

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
