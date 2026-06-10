import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const repoRoot = resolve(import.meta.dir, '../../../..')
const readRepoFile = (path: string) => readFileSync(resolve(repoRoot, path), 'utf8')

describe('nginx deployment ingress config', () => {
  test('proxies public and admin APIs from nginx', () => {
    const nginx = readRepoFile('web/nginx.conf')

    expect(nginx).toContain('proxy_pass http://api:8080')
    expect(nginx).toContain('proxy_pass http://api:9091')
    expect(nginx).not.toContain('return 308 /admin')
  })

  test('uses cloudflare origin certificate paths', () => {
    const nginx = readRepoFile('web/nginx.conf')

    expect(nginx).toContain('/etc/nginx/certs/cloudflare-origin.pem')
    expect(nginx).toContain('/etc/nginx/certs/cloudflare-origin.key')
  })

  test('production compose exposes web as the only public ingress', () => {
    const compose = readRepoFile('docker-compose.yaml')

    expect(compose).not.toContain('caddy:')
    expect(compose).toContain('"80:80"')
    expect(compose).toContain('"443:443"')
    expect(compose).toContain('./certs:/etc/nginx/certs:ro')
  })
})
