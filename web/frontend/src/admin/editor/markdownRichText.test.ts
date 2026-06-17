import { describe, expect, test } from 'bun:test'

import { looksLikeMarkdown } from './markdownDetection'

describe('markdown rich text helpers', () => {
  test('recognizes block and inline markdown authoring content', () => {
    expect(looksLikeMarkdown('# 标题\n\n正文')).toBe(true)
    expect(looksLikeMarkdown('- 第一项\n- 第二项')).toBe(true)
    expect(looksLikeMarkdown('这是一段 **加粗** 文本')).toBe(true)
    expect(looksLikeMarkdown('[链接](https://example.com)')).toBe(true)
  })

  test('keeps plain text and html out of the markdown path', () => {
    expect(looksLikeMarkdown('这是普通文本，版本是 1.0。')).toBe(false)
    expect(looksLikeMarkdown('API 返回 code: 14')).toBe(false)
    expect(looksLikeMarkdown('<p>已经是 HTML</p>')).toBe(false)
  })
})
