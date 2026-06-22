import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const src = (...parts: string[]) => resolve(import.meta.dir, '..', ...parts)
const read = (...parts: string[]) => readFileSync(src(...parts), 'utf8')

describe('admin category management source contract', () => {
  test('keeps paginated management API separate from full category selector API', () => {
    const api = read('admin/api/adminCategoryApi.ts')

    expect(api).toContain("adminHttp.get<AdminCategoryListResponse>('/categories'")
    expect(api).toContain("adminHttp.get<AdminCategoryAllResponse>('/categories/all')")
    expect(api).toContain('params')
  })

  test('management view uses paginated list state and keeps all-list out of the page', () => {
    const view = read('views/admin/AdminCategoryView.vue')

    expect(view).toContain('listAdminCategories')
    expect(view).not.toContain('listAllAdminCategories')
    expect(view).toContain('pageSize')
    expect(view).toContain('jumpPage')
    expect(view).toContain('totalPages')
  })

  test('management view exposes multi-select delete controls', () => {
    const view = read('views/admin/AdminCategoryView.vue')

    expect(view).toContain('selectedCategoryIds')
    expect(view).toContain('bulkDeleteCandidates')
    expect(view).toContain('confirmBulkDelete')
    expect(view).toContain('批量删除')
    expect(view).toContain(':disabled="category.is_system"')
  })

  test('public category list requests the paginated public endpoint', () => {
    const api = read('api/category.ts')
    const categoryList = read('components/category/CategoryList.vue')

    expect(api).toContain('ListCategoriesParams')
    expect(api).toContain("http.get('/categories'")
    expect(api).toContain('params')
    expect(categoryList).toContain('const categoryPageSize = 6')
    expect(categoryList).toContain('listCategories({ page: categoryPage.value, limit: categoryPageSize })')
    expect(categoryList).not.toContain('<PaginationControl')
    expect(categoryList).toContain('ChevronLeft')
    expect(categoryList).toContain('ChevronRight')
    expect(categoryList).toContain('aria-label="上一页"')
    expect(categoryList).toContain('aria-label="下一页"')
  })

  test('category deletion can request article deletion after explicit confirmation', () => {
    const api = read('admin/api/adminCategoryApi.ts')
    const view = read('views/admin/AdminCategoryView.vue')

    expect(api).toContain('delete_articles')
    expect(api).toContain('params: { delete_articles: deleteArticles }')
    expect(view).toContain('deleteArticlesWithCategory')
    expect(view).toContain('confirmDestructiveDelete')
    expect(view).toContain('同步删除分类下的文章')
  })
})
