import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const src = (...parts: string[]) => resolve(import.meta.dir, '..', ...parts)
const read = (...parts: string[]) => readFileSync(src(...parts), 'utf8')

describe('admin user management source contract', () => {
  test('registers backend users route and sidebar item', () => {
    const router = read('router/index.ts')
    const layout = read('views/admin/AdminLayout.vue')

    expect(router).toContain("name: 'adminUsers'")
    expect(router).toContain("path: 'users'")
    expect(layout).toContain("label: '用户'")
    expect(layout).toContain("activeRoutes: ['adminUsers']")
  })

  test('admin user API uses expected /v1 endpoints', () => {
    const api = read('admin/api/adminUserApi.ts')

    expect(api).toContain("adminHttp.get<AdminUserListResponse>('/users'")
    expect(api).toContain('`/users/${data.id}`')
    expect(api).toContain('`/users/${id}/disable`')
    expect(api).toContain('`/users/${id}/enable`')
  })

  test('management view exposes required controls', () => {
    const view = read('views/admin/AdminUserManagementView.vue')

    expect(view).toContain('用户管理')
    expect(view).toContain('placeholder="搜索用户名、姓名或邮箱"')
    expect(view).toContain('pageSize')
    expect(view).toContain('jumpPage')
    expect(view).toContain('selectedStatus')
    expect(view).toContain('openEdit')
    expect(view).toContain('openDisable')
    expect(view).toContain('openEnable')
  })
})
