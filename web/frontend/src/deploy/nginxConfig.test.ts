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

  test('api image build does not include production env secrets', () => {
    const dockerfile = readRepoFile('Dockerfile')
    const dockerignore = readRepoFile('.dockerignore')
    const deployWorkflow = readRepoFile('.github/workflows/deploy.yml')

    expect(dockerfile).not.toMatch(/COPY\s+--from=builder\s+\/app\/\.env\b/)
    expect(dockerignore).toMatch(/^\.env$/m)
    expect(dockerignore).toMatch(/^\.env\.\*$/m)
    expect(deployWorkflow).not.toContain('make decrypt_env env=prod')
  })

  test('compose injects api configuration at runtime instead of build time', () => {
    const productionCompose = readRepoFile('docker-compose.yaml')
    const developmentCompose = readRepoFile('docker-compose.dev.yaml')

    expect(productionCompose).toMatch(/api:\n[\s\S]*env_file:\n[\s\S]*path: \.env/)
    expect(developmentCompose).toMatch(/api:\n[\s\S]*env_file:\n[\s\S]*path: \.env/)
  })

  test('api and web images declare their runtime-only responsibilities', () => {
    const apiDockerfile = readRepoFile('Dockerfile')
    const webDockerfile = readRepoFile('web/Dockerfile')
    const webDevDockerfile = readRepoFile('web/Dockerfile.dev')

    expect(apiDockerfile).toContain('EXPOSE 8080 9091')
    expect(webDockerfile).not.toContain('openssl req -x509')
    expect(webDevDockerfile).toContain('openssl req -x509')
  })

  test('docs and env example do not list obsolete bootstrap variables', () => {
    const envExample = readRepoFile('.env.example')
    const readme = readRepoFile('README.md')
    const obsoleteKeys = [
      'DEFAULT_USER_ID',
      'DEFAULT_USERNAME',
      'DEFAULT_USER_PASSWORD',
      'DEFAULT_USER_FULLNAME',
      'DEFAULT_USER_EMAIL',
      'LETSENCRYPT_EMAIL',
    ]

    for (const key of obsoleteKeys) {
      expect(envExample).not.toContain(key)
      expect(readme).not.toContain(key)
    }
  })

  test('test workflow covers backend frontend and compose checks without env decryption', () => {
    const workflow = readRepoFile('.github/workflows/test.yml')

    expect(workflow).not.toContain('paths-ignore:')
    expect(workflow).not.toContain('make decrypt_env env=dev')
    expect(workflow).toContain('make test')
    expect(workflow).toContain('bun test')
    expect(workflow).toContain('bun run type-check')
    expect(workflow).toContain('bun run build')
    expect(workflow).toContain('docker compose config --quiet')
    expect(workflow).toContain('docker compose -f docker-compose.dev.yaml config --quiet')
  })
})
