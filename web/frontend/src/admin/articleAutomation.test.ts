import { describe, expect, test } from 'bun:test'
import type { AdminArticle } from './types'
import { isAutomationDraft } from './articleAutomation'

const baseArticle: AdminArticle = {
  id: 'article-1',
  title: 'Automation draft',
  summary: 'Summary',
  content: '<p>Content</p>',
  likes: 0,
  views: 0,
  is_publish: false,
  created_at: '2026-06-12T10:30:00Z',
  updated_at: '2026-06-12T10:30:00Z',
  owner: 'owner-1',
}

describe('isAutomationDraft', () => {
  test('matches unpublished automation drafts pending review', () => {
    expect(
      isAutomationDraft({
        ...baseArticle,
        created_by_automation: true,
        automation_status: 'pending_review',
      }),
    ).toBe(true)
  })

  test('does not match normal drafts or published automation articles', () => {
    expect(isAutomationDraft(baseArticle)).toBe(false)
    expect(
      isAutomationDraft({
        ...baseArticle,
        created_by_automation: true,
        automation_status: 'published',
      }),
    ).toBe(false)
    expect(
      isAutomationDraft({
        ...baseArticle,
        is_publish: true,
        created_by_automation: true,
        automation_status: 'pending_review',
      }),
    ).toBe(false)
  })
})
