# Post-Auth Stabilization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stabilize the auth-unification merge by documenting setup, cleaning legacy browser auth state on startup, and tightening admin login error handling.

**Architecture:** Keep the backend auth model unchanged. Move frontend storage key cleanup into a small testable helper, call it during auth store initialization, and update docs/config examples to match the unified setup flow.

**Tech Stack:** Vue 3, Pinia, Bun test runner, TypeScript, Go/Gin documentation and config examples.

---

## File Map

- Create: `web/frontend/src/store/module/authStorage.ts`
  - Owns auth storage key constants and legacy-key cleanup helpers.
- Create: `web/frontend/src/store/module/authStorage.test.ts`
  - Uses Bun's test runner with an in-memory storage implementation.
- Modify: `web/frontend/src/store/module/auth.ts`
  - Imports key constants and calls legacy-key cleanup before reading current auth state.
- Modify: `web/frontend/src/views/admin/AdminLoginView.vue`
  - Keeps role-denied and backend error extraction readable and safe.
- Modify: `web/frontend/package.json`
  - Adds a `test` script for `bun test`.
- Modify: `README.md`
  - Adds first-run setup and unified auth notes.
- Modify: `.env.example`
  - Clarifies current setup-related placeholder values.

---

### Task 1: Add Testable Auth Storage Cleanup

**Files:**
- Create: `web/frontend/src/store/module/authStorage.ts`
- Create: `web/frontend/src/store/module/authStorage.test.ts`
- Modify: `web/frontend/src/store/module/auth.ts`
- Modify: `web/frontend/package.json`

- [ ] **Step 1: Write the failing storage cleanup test**

Create `web/frontend/src/store/module/authStorage.test.ts`:

```ts
import { describe, expect, test } from 'bun:test'
import { AUTH_STORAGE_KEYS, LEGACY_AUTH_STORAGE_KEYS, cleanupLegacyAuthStorage } from './authStorage'

class MemoryStorage implements Pick<Storage, 'getItem' | 'setItem' | 'removeItem' | 'key' | 'length' | 'clear'> {
  private values = new Map<string, string>()

  get length() {
    return this.values.size
  }

  clear() {
    this.values.clear()
  }

  getItem(key: string) {
    return this.values.get(key) ?? null
  }

  key(index: number) {
    return Array.from(this.values.keys())[index] ?? null
  }

  removeItem(key: string) {
    this.values.delete(key)
  }

  setItem(key: string, value: string) {
    this.values.set(key, value)
  }
}

describe('cleanupLegacyAuthStorage', () => {
  test('removes legacy auth keys without removing current auth or unrelated keys', () => {
    const storage = new MemoryStorage()

    Object.values(AUTH_STORAGE_KEYS).forEach((key) => storage.setItem(key, `current:${key}`))
    LEGACY_AUTH_STORAGE_KEYS.forEach((key) => storage.setItem(key, `legacy:${key}`))
    storage.setItem('nostalgia-theme-mode', 'system')

    cleanupLegacyAuthStorage(storage)

    Object.values(AUTH_STORAGE_KEYS).forEach((key) => {
      expect(storage.getItem(key)).toBe(`current:${key}`)
    })
    LEGACY_AUTH_STORAGE_KEYS.forEach((key) => {
      expect(storage.getItem(key)).toBeNull()
    })
    expect(storage.getItem('nostalgia-theme-mode')).toBe('system')
  })
})
```

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```bash
cd web/frontend && bun test src/store/module/authStorage.test.ts
```

Expected: FAIL because `authStorage.ts` does not exist.

- [ ] **Step 3: Implement the storage helper**

Create `web/frontend/src/store/module/authStorage.ts`:

```ts
export const AUTH_STORAGE_KEYS = {
  TOKEN: 'nostalgia_user_token',
  TOKEN_EXPIRES: 'nostalgia_user_token_expires_at',
  REFRESH_TOKEN: 'nostalgia_user_refresh_token',
  REFRESH_TOKEN_EXPIRES: 'nostalgia_user_refresh_token_expires_at',
  USER: 'nostalgia_user_info',
} as const

export const LEGACY_AUTH_STORAGE_KEYS = [
  'nostalgia_access_token',
  'nostalgia_token_expires_at',
  'nostalgia_refresh_token',
  'nostalgia_refresh_token_expires_at',
  'nostalgia_admin_access_token',
  'nostalgia_admin_access_token_expires_at',
  'nostalgia_admin_refresh_token',
  'nostalgia_admin_refresh_token_expires_at',
  'nostalgia_admin_user',
] as const

export type AuthStorage = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>

export function cleanupLegacyAuthStorage(storage: AuthStorage = localStorage) {
  LEGACY_AUTH_STORAGE_KEYS.forEach((key) => storage.removeItem(key))
}
```

- [ ] **Step 4: Wire helper into the auth store**

Modify `web/frontend/src/store/module/auth.ts`:

```ts
import {
  AUTH_STORAGE_KEYS as STORAGE_KEYS,
  LEGACY_AUTH_STORAGE_KEYS,
  cleanupLegacyAuthStorage,
} from '@/store/module/authStorage'
```

Remove the inline `STORAGE_KEYS` and `LEGACY_STORAGE_KEYS` constants. Call cleanup before refs read from localStorage:

```ts
export const useAuthStore = defineStore('auth', () => {
  cleanupLegacyAuthStorage()

  const token = ref(localStorage.getItem(STORAGE_KEYS.TOKEN) || '')
```

Replace `LEGACY_STORAGE_KEYS` references with `LEGACY_AUTH_STORAGE_KEYS`.

- [ ] **Step 5: Add test script**

Modify `web/frontend/package.json`:

```json
"test": "bun test",
```

- [ ] **Step 6: Verify GREEN**

Run:

```bash
cd web/frontend && bun test src/store/module/authStorage.test.ts
cd web/frontend && bun run type-check
```

Expected: both commands exit 0.

- [ ] **Step 7: Commit**

```bash
git add web/frontend/package.json web/frontend/src/store/module/auth.ts web/frontend/src/store/module/authStorage.ts web/frontend/src/store/module/authStorage.test.ts
git commit -m "fix(frontend): clear legacy auth storage on startup"
```

---

### Task 2: Refresh Setup And Auth Documentation

**Files:**
- Modify: `README.md`
- Modify: `.env.example`

- [ ] **Step 1: Update README setup flow**

Add a first-run setup section that shows:

```markdown
### 首次初始化管理员

1. 在 `.env` 中设置 `TOKEN_SYMMETRIC_KEY` 和 `SETUP_TOKEN`。
2. 运行数据库迁移。
3. 启动服务后访问 `/setup`。
4. 使用 `SETUP_TOKEN` 创建第一个 `admin` 用户。
5. 后续进入 `/admin/login` 使用该账号管理内容。
```

Also mention that public registration creates only `visitor` users.

- [ ] **Step 2: Update `.env.example` placeholders**

Keep `SETUP_TOKEN=` present and document by value shape only:

```env
SETUP_TOKEN=replace-with-a-random-one-time-bootstrap-token
```

Do not add real secrets.

- [ ] **Step 3: Verify docs diff**

Run:

```bash
git diff -- README.md .env.example
```

Expected: documentation only, no secrets.

- [ ] **Step 4: Commit**

```bash
git add README.md .env.example
git commit -m "docs(auth): document setup stabilization flow"
```

---

### Task 3: Final Verification

**Files:**
- No planned source edits.

- [ ] **Step 1: Run frontend verification**

Run:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

Expected: all commands exit 0.

- [ ] **Step 2: Run backend verification**

Run:

```bash
make test
```

Expected: exits 0.

- [ ] **Step 3: Review branch state**

Run:

```bash
git status --short --branch
git log --oneline --decorate --max-count=8
```

Expected: branch is `chore/post-auth-stabilization`; working tree is clean after commits.
