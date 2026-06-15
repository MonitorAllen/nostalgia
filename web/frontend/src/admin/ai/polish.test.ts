import { describe, expect, test } from 'bun:test'
import {
  buildAIPolishRequest,
  createAIPolishSession,
  getAIPolishApplyLabel,
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

  test('creates pending content polish sessions', () => {
    expect(
      createAIPolishSession({
        mode: 'improve',
        target: 'content_selection',
        sourceText: '原文'
      })
    ).toMatchObject({
      mode: 'improve',
      target: 'content_selection',
      sourceText: '原文',
      status: 'loading',
      selectedSuggestionIndex: -1
    })
  })

  test('labels explicit apply actions', () => {
    expect(getAIPolishApplyLabel('content_selection')).toBe('替换选区')
    expect(getAIPolishApplyLabel('title')).toBe('应用到标题')
    expect(getAIPolishApplyLabel('summary')).toBe('应用到摘要')
  })
})
