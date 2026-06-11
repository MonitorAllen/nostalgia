import { describe, expect, test } from 'bun:test'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative, resolve } from 'node:path'

const frontendSrcRoot = resolve(import.meta.dir, '..')
const ignoredDirectories = new Set(['node_modules', 'dist'])

const sourceFiles = (directory: string): string[] =>
  readdirSync(directory).flatMap((entry) => {
    const path = join(directory, entry)
    const stats = statSync(path)

    if (stats.isDirectory()) {
      if (ignoredDirectories.has(entry)) return []
      return sourceFiles(path)
    }

    return /\.(ts|vue)$/.test(entry) ? [path] : []
  })

describe('frontend source hygiene', () => {
  test('does not ship stray console logging in production source', () => {
    const offenders = sourceFiles(frontendSrcRoot)
      .map((path) => ({
        path: relative(frontendSrcRoot, path),
        source: readFileSync(path, 'utf8')
      }))
      .filter(({ path }) => !path.endsWith('.test.ts'))
      .flatMap(({ path, source }) =>
        [...source.matchAll(/\bconsole\.(log|debug|info)\s*\(/g)].map((match) => `${path}:${match[1]}`)
      )

    expect(offenders).toEqual([])
  })
})
