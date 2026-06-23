import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const src = (...parts: string[]) => resolve(import.meta.dir, '..', '..', ...parts)
const read = (...parts: string[]) => readFileSync(src(...parts), 'utf8')

describe('article cover rendering contracts', () => {
  test('shared ArticleCover component uses a stable 16:9 object-cover container', () => {
    const component = read('components/article/ArticleCover.vue')

    expect(component).toContain('aspect-[16/9]')
    expect(component).toContain('object-cover')
    expect(component).toContain('fallbackSrc')
    expect(component).toContain('handleImageError')
  })
})
