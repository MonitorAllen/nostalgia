import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const src = (...parts: string[]) => resolve(import.meta.dir, '..', '..', ...parts)
const read = (...parts: string[]) => readFileSync(src(...parts), 'utf8')

describe('article cover rendering contracts', () => {
  test('shared ArticleCover component uses a stable 16:9 object-cover container', () => {
    const component = read('components/article/ArticleCover.vue')

    expect(component).toContain('<span :class="containerClass">')
    expect(component).toContain('<img v-if="displaySrc"')
    expect(component).toContain('aspect-[16/9]')
    expect(component).toContain('object-cover')
    expect(component).toContain('fallbackSrc')
    expect(component).toContain('handleImageError')
    expect(component).toContain("if (props.variant === 'list') return base")
    expect(component).toContain('transition-transform duration-200')
    expect(component).not.toContain('return `${base} h-full`')
  })

  test('ArticleReader renders the cover before the header when showCover is enabled', () => {
    const reader = read('components/article/ArticleReader.vue')

    expect(reader).toContain("import ArticleCover from './ArticleCover.vue'")
    expect(reader).toContain('<ArticleCover')
    expect(reader.indexOf('<ArticleCover')).toBeLessThan(reader.indexOf('<header'))
    expect(reader).toContain(':src="cover"')
    expect(reader).toContain('variant="detail"')
  })

  test('public article detail passes cover into ArticleReader', () => {
    const view = read('views/article/ArticleView.vue')

    expect(view).toContain(':cover="article.cover"')
    expect(view).toContain('show-cover')
  })

  test('ArticleList uses shared object-cover rendering without padded contain thumbnails', () => {
    const list = read('components/article/ArticleList.vue')

    expect(list).toContain("import ArticleCover from './ArticleCover.vue'")
    expect(list).toContain('variant="list"')
    expect(list).toContain('fallback-src="/images/go.png"')
    expect(list).toContain('md:self-start')
    expect(list).not.toContain('object-contain')
    expect(list).not.toContain('p-2 transition')
    expect(list).not.toContain('md:h-full')
  })
})
