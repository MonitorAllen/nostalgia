import { describe, expect, test } from 'bun:test'
import {
  buildAIPolishRequest,
  buildAIPolishContentPreview,
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

  test('builds rich content selection requests without changing plain text fallback', () => {
    expect(
      buildAIPolishRequest({
        mode: 'improve',
        target: 'content_selection',
        text: '推荐的配置分层',
        inputFormat: 'html',
        richText: '<h2>推荐的配置分层</h2><ul><li>本地开发</li></ul>'
      })
    ).toMatchObject({
      mode: 'improve',
      target: 'content_selection',
      text: '推荐的配置分层',
      input_format: 'html',
      rich_text: '<h2>推荐的配置分层</h2><ul><li>本地开发</li></ul>'
    })
  })

  test('previews rich content selections by replacing the selected html fragment', () => {
    expect(
      buildAIPolishContentPreview({
        articleContent:
          '<h2>推荐的配置分层</h2><ul><li>本地开发</li><li>线上环境</li></ul><p>后续说明</p>',
        sourceText: '推荐的配置分层\n本地开发\n线上环境',
        sourceRichText: '<h2>推荐的配置分层</h2><ul><li>本地开发</li><li>线上环境</li></ul>',
        replacementHtml:
          '<h2>配置分层建议</h2><table><tbody><tr><td>本地</td><td>快速验证</td></tr></tbody></table>'
      })
    ).toBe(
      '<h2>配置分层建议</h2><table><tbody><tr><td>本地</td><td>快速验证</td></tr></tbody></table><p>后续说明</p>'
    )
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
