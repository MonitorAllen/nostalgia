import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const repoRoot = resolve(import.meta.dir, '../../../..')
const readRepoFile = (path: string) => readFileSync(resolve(repoRoot, path), 'utf8')

const legacyTables = ['admins', 'roles', 'role_permissions', 'sys_menus']

describe('database legacy structure audit', () => {
  test('current database documentation does not expose removed admin rbac tables', () => {
    const schema = readRepoFile('doc/schema.sql')
    const dbml = readRepoFile('doc/db.dbml')

    for (const table of legacyTables) {
      expect(schema).not.toContain(`CREATE TABLE "${table}"`)
      expect(dbml).not.toContain(`Table "${table}"`)
    }
  })

  test('frontend admin API types no longer model legacy role ids', () => {
    const adminTypes = readRepoFile('web/frontend/src/admin/types.ts')

    expect(adminTypes).not.toContain('role_id')
    expect(adminTypes).toContain('role?: string')
  })

  test('runtime sqlc models do not include legacy admin rbac tables', () => {
    const models = readRepoFile('db/sqlc/models.go')

    expect(models).not.toContain('type Admin struct')
    expect(models).not.toContain('type Role struct')
    expect(models).not.toContain('type RolePermission struct')
    expect(models).not.toContain('type SysMenu struct')
  })
})
