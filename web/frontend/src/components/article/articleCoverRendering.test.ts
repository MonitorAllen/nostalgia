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
})
