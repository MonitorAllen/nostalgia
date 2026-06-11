import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const repoRoot = resolve(import.meta.dir, '../../../..')
const readRepoFile = (path: string) => readFileSync(resolve(repoRoot, path), 'utf8')

describe('vite bundle splitting config', () => {
  test('keeps editor and code highlighting dependencies out of the public entry chunk', () => {
    const config = readRepoFile('web/frontend/vite.config.ts')

    expect(config).toContain('manualChunks')
    expect(config).toContain('ckeditor')
    expect(config).toContain('prism')
    expect(config).toContain('content-rendering')
    expect(config).toContain('admin-editor')
  })

  test('lazy-loads article detail so Prism is not part of the app shell route table', () => {
    const router = readRepoFile('web/frontend/src/router/index.ts')

    expect(router).not.toContain("import ArticleView from '@/views/article/ArticleView.vue'")
    expect(router).toContain("component: () => import('@/views/article/ArticleView.vue')")
  })
})
