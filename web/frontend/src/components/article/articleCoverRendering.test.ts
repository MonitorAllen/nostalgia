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

  test('public article list presents covers as medium-emphasis reading cards', () => {
    const list = read('components/article/ArticleList.vue')

    expect(list).toContain("import ArticleCover from './ArticleCover.vue'")
    expect(list).toContain('md:grid-cols-[15rem_minmax(0,1fr)]')
    expect(list).toContain('variant="list"')
    expect(list).toContain('fallback-src="/images/go.png"')
    expect(list).toContain('md:aspect-[16/9]')
    expect(list).toContain('group-hover:border-accent/35')
    expect(list).not.toContain('md:grid-cols-[13rem_1fr]')
  })

  test('admin article list uses shared covers and previews from cover clicks', () => {
    const view = read('views/admin/AdminArticleListView.vue')

    expect(view).toContain("import ArticleCover from '@/components/article/ArticleCover.vue'")
    expect(view).toContain('<ArticleCover')
    expect(view).toContain(':src="article.cover"')
    expect(view).toContain('fallback-src="/images/go.png"')
    expect(view).toContain('@click="openArticlePreview(article)"')
    expect(view).toContain(":aria-label=\"`预览 ${article.title || '无标题文章'} 封面`\"")
    expect(view).not.toContain('coverLabel(article)')
    expect(view).not.toContain('@click="editArticle(article.id)"\n            >\n              <img')
  })

  test('admin cover panel exposes guidance, warnings, and multi-surface previews', () => {
    const panel = read('admin/editor/AdminArticleCoverPanel.vue')

    expect(panel).toContain('推荐 1600x900')
    expect(panel).toContain('1920x1080')
    expect(panel).toContain('详情页头图')
    expect(panel).toContain('列表卡片')
    expect(panel).toContain('分享预览')
    expect(panel).toContain('inspection.warnings')
    expect(panel).toContain('ArticleCover')
  })

  test('admin editor wires cover diagnostics into the cover panel', () => {
    const view = read('views/admin/AdminArticleEditorView.vue')

    expect(view).toContain("import AdminArticleCoverPanel from '@/admin/editor/AdminArticleCoverPanel.vue'")
    expect(view).toContain('inspectArticleCoverDimensions')
    expect(view).toContain('loadArticleCoverDimensions')
    expect(view).toContain('loadArticleCoverFileDimensions')
    expect(view).toContain('coverInspection')
    expect(view).toContain('<AdminArticleCoverPanel')
    expect(view).toContain(':inspection="coverInspection"')
    expect(view).toContain('const requestToken = ++coverInspectionToken')
    expect(view).toContain('if (requestToken !== coverInspectionToken) return')
    expect(view).toContain('coverInspection.value = null')
    expect(view.indexOf('const requestToken = ++coverInspectionToken')).toBeLessThan(
      view.indexOf('if (requestToken !== coverInspectionToken) return')
    )
    expect(view.indexOf('void inspectCoverFromFile(file!)')).toBeLessThan(
      view.indexOf('if (!article.value.id) return')
    )
    expect(view).toContain("const coverInspection = ref<ArticleCoverInspection | null>(null)")
  })
})
