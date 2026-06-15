import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const repoRoot = resolve(import.meta.dir, '../../../..')
const readRepoFile = (path: string) => readFileSync(resolve(repoRoot, path), 'utf8')

describe('frontend SEO integration', () => {
  test('index html ships crawl-safe baseline metadata before Vue mounts', () => {
    const indexHtml = readRepoFile('web/frontend/index.html')

    expect(indexHtml).toContain('<html lang="zh-CN">')
    expect(indexHtml).toContain('<meta name="description"')
    expect(indexHtml).toContain('<meta name="robots" content="index,follow">')
    expect(indexHtml).toContain('<meta property="og:site_name" content="Nostalgia">')
  })

  test('router applies static metadata for every navigation', () => {
    const router = readRepoFile('web/frontend/src/router/index.ts')

    expect(router).toContain("import { applySeoMetadata, buildRouteSeoMetadata } from '@/util/seo'")
    expect(router).toContain('router.afterEach')
    expect(router).toContain('applySeoMetadata(buildRouteSeoMetadata(to))')
  })

  test('article detail applies article-specific canonical metadata after load', () => {
    const articleView = readRepoFile('web/frontend/src/views/article/ArticleView.vue')

    expect(articleView).toContain("import { applySeoMetadata, buildArticleSeoMetadata } from '@/util/seo'")
    expect(articleView).toContain('applySeoMetadata(buildArticleSeoMetadata(article.value))')
  })

  test('category and search views refine route metadata from visible page context', () => {
    const categoryView = readRepoFile('web/frontend/src/views/category/CategoryArticleView.vue')
    const searchView = readRepoFile('web/frontend/src/views/article/SearchArticleView.vue')

    expect(categoryView).toContain("import { applySeoMetadata, buildCategorySeoMetadata } from '@/util/seo'")
    expect(categoryView).toContain('applySeoMetadata(buildCategorySeoMetadata')
    expect(searchView).toContain("import { applySeoMetadata, buildSearchSeoMetadata } from '@/util/seo'")
    expect(searchView).toContain('applySeoMetadata(buildSearchSeoMetadata')
  })
})
