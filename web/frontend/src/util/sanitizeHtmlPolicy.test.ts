import { describe, expect, test } from 'bun:test'

import {
  getSafeTargetRel,
  getSanitizedHtmlConfig,
  isAllowedSanitizedUri,
  SANITIZED_HTML_CONFIG
} from './sanitizeHtmlPolicy'

describe('sanitizeHtmlPolicy', () => {
  test('allows only public article-safe URL protocols and same-origin references', () => {
    expect(isAllowedSanitizedUri('https://example.com/post')).toBe(true)
    expect(isAllowedSanitizedUri('http://example.com/post')).toBe(true)
    expect(isAllowedSanitizedUri('mailto:author@example.com')).toBe(true)
    expect(isAllowedSanitizedUri('tel:+15551234567')).toBe(true)
    expect(isAllowedSanitizedUri('/resources/content/image.png')).toBe(true)
    expect(isAllowedSanitizedUri('#comments')).toBe(true)
  })

  test('rejects scriptable or embedded-data URL protocols', () => {
    expect(isAllowedSanitizedUri('javascript:alert(1)')).toBe(false)
    expect(isAllowedSanitizedUri(' java\nscript:alert(1)')).toBe(false)
    expect(isAllowedSanitizedUri('data:text/html,<script>alert(1)</script>')).toBe(false)
    expect(isAllowedSanitizedUri('vbscript:msgbox(1)')).toBe(false)
    expect(isAllowedSanitizedUri('ftp://example.com/file')).toBe(false)
  })

  test('normalizes target links to prevent opener access', () => {
    expect(getSafeTargetRel()).toBe('noopener noreferrer')
    expect(getSafeTargetRel('nofollow')).toBe('nofollow noopener noreferrer')
    expect(getSafeTargetRel('noopener')).toBe('noopener noreferrer')
  })

  test('keeps the base DOMPurify config focused on CKEditor output', () => {
    expect(SANITIZED_HTML_CONFIG.ADD_ATTR).toContain('target')
    expect(SANITIZED_HTML_CONFIG.FORBID_TAGS).toEqual(['script', 'style', 'iframe', 'object', 'embed'])
  })

  test('keeps article output visually aligned while making comments stricter', () => {
    expect(getSanitizedHtmlConfig('article').FORBID_ATTR ?? []).not.toContain('style')
    expect(getSanitizedHtmlConfig('comment').FORBID_ATTR).toContain('style')
  })
})
