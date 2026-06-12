import { describe, expect, test } from 'bun:test'

import {
  buildArticleSeoMetadata,
  buildRouteSeoMetadata,
  createSeoHeadDescriptors,
  getCanonicalArticlePath,
  resolveSiteOrigin,
  stripHtmlToPlainText,
  toAbsoluteUrl
} from './seo'
import type { Article } from '@/types/article'

const article: Article = {
  id: '6f1d2ad1-07e8-490f-8f70-4ec4128b58c9',
  title: 'Redis 缓存一致性实践',
  summary: '从分页缓存、版本化 key 和失效策略聊起。',
  content: '<p>缓存不是把数据塞进 Redis 就结束了。</p><script>alert("x")</script>',
  views: 8,
  likes: 3,
  is_publish: true,
  cover: '/resources/covers/redis.png',
  slug: 'redis-cache-consistency',
  check_outdated: false,
  last_updated: '2026-06-10T08:00:00Z',
  read_time: '8 min',
  owner: 'owner-id',
  created_at: '2026-06-01T08:00:00Z',
  updated_at: '2026-06-11T08:00:00Z',
  deleted_at: '0001-01-01T00:00:00Z',
  username: 'allen',
  category_name: 'Go'
}

describe('seo helpers', () => {
  test('normalizes canonical origins and absolute URLs', () => {
    expect(resolveSiteOrigin('https://example.com/')).toBe('https://example.com')
    expect(resolveSiteOrigin('', 'https://fallback.example')).toBe('https://fallback.example')
    expect(toAbsoluteUrl('/article/redis', 'https://example.com/base/')).toBe(
      'https://example.com/article/redis'
    )
    expect(toAbsoluteUrl('https://cdn.example.com/cover.png', 'https://example.com')).toBe(
      'https://cdn.example.com/cover.png'
    )
  })

  test('extracts stable plain text for metadata descriptions', () => {
    expect(stripHtmlToPlainText('<p>Hello&nbsp;<strong>world</strong></p>')).toBe('Hello world')
    expect(stripHtmlToPlainText('<style>.x{}</style><script>alert(1)</script><p>正文</p>')).toBe(
      '正文'
    )
  })

  test('builds article canonical URLs from slugs before ids', () => {
    expect(getCanonicalArticlePath(article)).toBe('/article/redis-cache-consistency')
    expect(getCanonicalArticlePath({ ...article, slug: '' })).toBe(
      '/article/6f1d2ad1-07e8-490f-8f70-4ec4128b58c9'
    )
  })

  test('builds article metadata and JSON-LD from rendered content fields', () => {
    const metadata = buildArticleSeoMetadata(article, { siteOrigin: 'https://blog.example.com/' })

    expect(metadata.title).toBe('Redis 缓存一致性实践 | Nostalgia')
    expect(metadata.description).toBe('从分页缓存、版本化 key 和失效策略聊起。')
    expect(metadata.canonicalUrl).toBe('https://blog.example.com/article/redis-cache-consistency')
    expect(metadata.robots).toBe('index,follow')
    expect(metadata.openGraph?.type).toBe('article')
    expect(metadata.openGraph?.image).toBe('https://blog.example.com/resources/covers/redis.png')
    expect(metadata.jsonLd).toMatchObject({
      '@context': 'https://schema.org',
      '@type': 'BlogPosting',
      headline: 'Redis 缓存一致性实践',
      author: {
        '@type': 'Person',
        name: 'allen'
      },
      datePublished: '2026-06-01T08:00:00Z',
      dateModified: '2026-06-11T08:00:00Z',
      mainEntityOfPage: 'https://blog.example.com/article/redis-cache-consistency'
    })
  })

  test('marks private routes noindex and public routes indexable', () => {
    expect(buildRouteSeoMetadata({ name: 'adminLogin', path: '/backend/login' }).robots).toBe(
      'noindex,nofollow'
    )
    expect(buildRouteSeoMetadata({ name: 'home', path: '/' })).toMatchObject({
      title: 'Nostalgia | 技术文章与开发笔记',
      canonicalPath: '/',
      robots: 'index,follow'
    })
  })

  test('creates one deterministic descriptor per managed head target', () => {
    const metadata = buildArticleSeoMetadata(article, { siteOrigin: 'https://blog.example.com' })
    const descriptors = createSeoHeadDescriptors(metadata)

    expect(descriptors).toContainEqual({
      kind: 'title',
      value: 'Redis 缓存一致性实践 | Nostalgia'
    })
    expect(descriptors).toContainEqual({
      kind: 'meta',
      key: 'description',
      keyAttr: 'name',
      content: '从分页缓存、版本化 key 和失效策略聊起。'
    })
    expect(descriptors).toContainEqual({
      kind: 'link',
      rel: 'canonical',
      href: 'https://blog.example.com/article/redis-cache-consistency'
    })
    expect(descriptors.filter((descriptor) => descriptor.kind === 'jsonLd')).toHaveLength(1)
  })
})
