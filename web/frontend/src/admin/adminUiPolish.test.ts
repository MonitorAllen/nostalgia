import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const src = resolve(import.meta.dir, '..')

const readSource = (path: string) => readFileSync(resolve(src, path), 'utf8')

describe('admin UI polish contracts', () => {
  test('admin shell uses a collapsible flush sidebar with tag identity and AI settings nav', () => {
    const source = readSource('views/admin/AdminLayout.vue')

    expect(source).toContain('isSidebarCollapsed')
    expect(source).toContain('isEditorRoute')
    expect(source).toContain('PanelLeftClose')
    expect(source).toContain('Tags')
    expect(source).toContain("name: 'adminAiSettings'")
    expect(source).toContain('lg:grid-cols-[var(--admin-sidebar-width)_minmax(0,1fr)]')
  })

  test('admin login uses a tag icon instead of the public logo image', () => {
    const source = readSource('views/admin/AdminLoginView.vue')

    expect(source).toContain("import { Tags }")
    expect(source).toContain('<Tags')
    expect(source).not.toContain('src="/logo.svg"')
  })

  test('article list keeps automation drafts first and renders cover thumbnails', () => {
    const source = readSource('views/admin/AdminArticleListView.vue')

    expect(source).toContain('visibleArticles')
    expect(source).toContain('sortAutomationDraftsFirst')
    expect(source).toContain('coverLabel')
    expect(source).toContain(':src="coverLabel(article)"')
    expect(source).toContain('aria-label="搜索文章"')
    expect(source).not.toContain('快速检查文章状态、调整发布节奏')
  })

  test('editor moves title into settings and exposes article preview modal', () => {
    const source = readSource('views/admin/AdminArticleEditorView.vue')

    expect(source).toContain('previewOpen')
    expect(source).toContain('openPreview')
    expect(source).toContain('sanitizeHtml')
    expect(source).toContain('文章预览')
    expect(source).toContain('实际阅读效果')
    expect(source).toContain('aria-label="文章标题"')
    expect(source).not.toContain('class="h-12 rounded-archive text-base font-black sm:text-lg"')
  })

  test('admin routes include a dedicated AI settings page', () => {
    const routerSource = readSource('router/index.ts')
    const pageSource = readSource('views/admin/AdminAISettingsView.vue')
    const apiSource = readSource('admin/api/adminAiApi.ts')

    expect(routerSource).toContain("name: 'adminAiSettings'")
    expect(pageSource).toContain('AI 提供商')
    expect(pageSource).toContain('getAdminAIConfig')
    expect(apiSource).toContain('getAdminAIConfig')
  })
})

describe('public navigation polish contracts', () => {
  test('profile menu has outside click close and an admin backend entry', () => {
    const source = readSource('views/layout/NavBar.vue')

    expect(source).toContain('userMenuRef')
    expect(source).toContain('handleDocumentClick')
    expect(source).toContain('onBeforeUnmount')
    expect(source).toContain("userStore.userInfo?.role === 'admin'")
    expect(source).toContain("name: 'adminArticles'")
    expect(source).toContain('后台管理')
  })

  test('theme switching animates global color changes and icon controls', () => {
    const useThemeSource = readSource('composables/useTheme.ts')
    const cssSource = readSource('assets/main.css')
    const switcherSource = readSource('components/ui/ThemeSwitcher.vue')

    expect(useThemeSource).toContain('THEME_TRANSITION_CLASS')
    expect(useThemeSource).toContain('theme-transitioning')
    expect(cssSource).toContain('html.theme-transitioning')
    expect(switcherSource).toContain('duration-200')
    expect(switcherSource).toContain('scale-105')
  })
})
