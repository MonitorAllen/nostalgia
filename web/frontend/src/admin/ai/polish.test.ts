import { describe, expect, test } from 'bun:test'
import {
  buildAIPolishRequest,
  getAIPolishModeLabel,
  normalizeSelectedText,
  truncateForAIPolish
} from './polish'

describe('AI polish helpers', () => {
  test('normalizes selected text', () => {
    expect(normalizeSelectedText('  hello\n\n world  ')).toBe('hello\nworld')
  })

  test('truncates long context', () => {
    expect(truncateForAIPolish('abcdef', 4)).toBe('abcd')
  })

  test('builds content selection requests', () => {
    expect(
      buildAIPolishRequest({
        mode: 'improve',
        target: 'content_selection',
        text: 'hello',
        articleTitle: 'Title',
        articleSummary: 'Summary',
        articleExcerpt: 'Excerpt'
      })
    ).toEqual({
      mode: 'improve',
      target: 'content_selection',
      text: 'hello',
      article_title: 'Title',
      article_summary: 'Summary',
      article_excerpt: 'Excerpt',
      locale: 'zh-CN'
    })
  })

  test('builds title candidate requests with bounded excerpt', () => {
    expect(
      buildAIPolishRequest({
        mode: 'title_candidates',
        target: 'title',
        text: 'Current title',
        articleSummary: 'Summary',
        articleExcerpt: 'x'.repeat(12),
        maxContextChars: 8
      }).article_excerpt
    ).toBe('xxxxxxxx')
  })

  test('labels modes', () => {
    expect(getAIPolishModeLabel('shorten')).toBe('精简')
    expect(getAIPolishModeLabel('summary_candidates')).toBe('摘要候选')
  })
})
