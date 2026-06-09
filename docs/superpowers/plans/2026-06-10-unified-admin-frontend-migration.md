# Unified Admin Frontend Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an owner-only `/admin` experience inside `web/frontend` and retire the legacy Geeker Admin frontend after the migrated workflows are verified.

**Architecture:** Keep public blog routes and APIs intact while adding a separate admin route group, admin auth store, and `/v1` admin API client inside `web/frontend`. The admin UI reuses the glass/archive design system and only migrates owner-needed article and category workflows.

**Tech Stack:** Vue 3, Vite, TypeScript, Pinia, Vue Router, Tailwind CSS, Reka UI, lucide icons, CKEditor 5, Bun, gRPC-Gateway `/v1`, Gin `/api`.

---

## File Structure

- Create `web/frontend/src/admin/types.ts`: shared admin, article, category, list, token, and upload response types.
- Create `web/frontend/src/admin/api/adminHttp.ts`: axios wrapper for `/v1`, admin auth headers, refresh retry, and toast errors.
- Create `web/frontend/src/admin/api/adminAuthApi.ts`: admin login, token refresh, and admin info calls.
- Create `web/frontend/src/admin/api/adminArticleApi.ts`: admin article CRUD calls.
- Create `web/frontend/src/admin/api/adminCategoryApi.ts`: admin category CRUD calls.
- Create `web/frontend/src/admin/api/adminUploadApi.ts`: `/v1/util/upload_file` upload call.
- Create `web/frontend/src/admin/stores/adminAuth.ts`: dedicated Pinia store using `nostalgia_admin_*` localStorage keys.
- Create `web/frontend/src/views/admin/AdminLayout.vue`: admin shell with sidebar/topbar, theme switcher, site link, and logout.
- Create `web/frontend/src/views/admin/AdminLoginView.vue`: owner login form.
- Create `web/frontend/src/views/admin/AdminArticleListView.vue`: article list, search, paging, publish toggle, and delete.
- Create `web/frontend/src/views/admin/AdminArticleEditorView.vue`: CKEditor article editor with metadata, draft cache, upload, and save status.
- Create `web/frontend/src/views/admin/AdminCategoryView.vue`: category list, create, rename, and delete.
- Create `web/frontend/src/admin/editor/adminEditorConfig.ts`: frontend-local CKEditor config derived from old admin config.
- Create `web/frontend/src/admin/editor/adminUploadAdapter.ts`: CKEditor upload adapter using `adminUploadApi`.
- Modify `web/frontend/src/router/index.ts`: add `/admin` routes and route guard.
- Modify `web/frontend/src/App.vue`: hide public footer for admin routes.
- Modify `web/frontend/vite.config.ts`: add `/v1` proxy.
- Modify `web/frontend/src/assets/content.css`: add admin editor content bridge styles when needed.
- Modify `AGENTS.md`: document the unified frontend admin after cleanup.
- Delete `web/backend/` after migrated workflows build successfully.

## Task 1: Admin API Client, Types, Auth Store, and Dev Proxy

**Files:**
- Create: `web/frontend/src/admin/types.ts`
- Create: `web/frontend/src/admin/api/adminHttp.ts`
- Create: `web/frontend/src/admin/api/adminAuthApi.ts`
- Create: `web/frontend/src/admin/api/adminArticleApi.ts`
- Create: `web/frontend/src/admin/api/adminCategoryApi.ts`
- Create: `web/frontend/src/admin/api/adminUploadApi.ts`
- Create: `web/frontend/src/admin/stores/adminAuth.ts`
- Modify: `web/frontend/vite.config.ts`

- [ ] **Step 1: Add `/v1` proxy to the frontend dev server**

Edit `web/frontend/vite.config.ts` so `server.proxy` contains this exact entry while keeping the existing `/api` and `/resources` entries:

```ts
proxy: {
  '/api': 'http://localhost:8080',
  '/v1': {
    target: 'http://localhost:9091',
    changeOrigin: true,
  },
  '/resources': 'http://localhost:8080'
}
```

- [ ] **Step 2: Create shared admin types**

Create `web/frontend/src/admin/types.ts`:

```ts
export interface AdminUser {
  id: number
  username: string
  is_active?: boolean
  role_id?: number
  created_at?: string
}

export interface AdminTokens {
  access_token: string
  access_token_expires_at: string
  refresh_token: string
  refresh_token_expires_at: string
}

export interface AdminLoginRequest {
  username: string
  password: string
}

export interface AdminLoginResponse extends AdminTokens {
  admin: AdminUser
}

export interface AdminArticle {
  id: string
  title: string
  summary: string
  content: string
  likes: number
  views: number
  is_publish: boolean
  created_at: string
  updated_at: string
  owner: string
  category_id?: number
  category_name?: string
  cover?: string
  slug?: string
  check_outdated?: boolean
}

export interface AdminArticleListResponse {
  articles: AdminArticle[]
  count: string | number
}

export interface AdminArticleResponse {
  article: AdminArticle
}

export interface AdminCategory {
  id: number
  name: string
  article_count?: number
  created_at?: string
  updated_at?: string
}

export interface AdminCategoryListResponse {
  categories: AdminCategory[]
  count: string | number
}

export interface AdminUploadRequest {
  article_id: string
  content: string
  type: 'content' | 'cover'
}

export interface AdminUploadResponse {
  url: string
  filename: string
}
```

- [ ] **Step 3: Create admin auth API module**

Create `web/frontend/src/admin/api/adminAuthApi.ts`:

```ts
import adminHttp from './adminHttp'
import type { AdminLoginRequest, AdminLoginResponse, AdminTokens, AdminUser } from '../types'

export function loginAdmin(data: AdminLoginRequest) {
  return adminHttp.post<AdminLoginResponse>('/admin/login', data, {
    skipAuth: true,
    skipErrorHandler: true,
  })
}

export function renewAdminAccessToken(refreshToken: string) {
  return adminHttp.post<Pick<AdminTokens, 'access_token' | 'access_token_expires_at'>>(
    '/admin/renew_access',
    { refresh_token: refreshToken },
    { skipAuth: true, skipErrorHandler: true },
  )
}

export function getAdminInfo() {
  return adminHttp.get<{ admin: AdminUser }>('/admin/info')
}
```

- [ ] **Step 4: Create admin article API module**

Create `web/frontend/src/admin/api/adminArticleApi.ts`:

```ts
import adminHttp from './adminHttp'
import type { AdminArticle, AdminArticleListResponse, AdminArticleResponse } from '../types'

export interface ListAdminArticlesParams {
  title?: string
  page: number
  limit: number
}

export function listAdminArticles(params: ListAdminArticlesParams) {
  return adminHttp.get<AdminArticleListResponse>('/articles', { params })
}

export function getAdminArticle(id: string, needContent = true) {
  return adminHttp.get<AdminArticleResponse>(`/articles/${id}/${needContent}`)
}

export function createAdminArticle(data: Partial<AdminArticle>) {
  return adminHttp.post<AdminArticleResponse>('/articles', data)
}

export function updateAdminArticle(data: Partial<AdminArticle>) {
  return adminHttp.patch<AdminArticleResponse>('/articles', data)
}

export function deleteAdminArticle(id: string) {
  return adminHttp.delete<void>(`/articles/${id}`)
}
```

- [ ] **Step 5: Create admin category and upload API modules**

Create `web/frontend/src/admin/api/adminCategoryApi.ts`:

```ts
import adminHttp from './adminHttp'
import type { AdminCategory, AdminCategoryListResponse } from '../types'

export function listAllAdminCategories() {
  return adminHttp.get<AdminCategoryListResponse>('/categories/all')
}

export function createAdminCategory(data: { name: string }) {
  return adminHttp.post<{ category: AdminCategory }>('/categories', data)
}

export function updateAdminCategory(data: { id: number; name: string }) {
  return adminHttp.patch<{ category: AdminCategory }>('/categories', data)
}

export function deleteAdminCategory(id: number) {
  return adminHttp.delete<void>(`/categories/${id}`)
}
```

Create `web/frontend/src/admin/api/adminUploadApi.ts`:

```ts
import adminHttp from './adminHttp'
import type { AdminUploadRequest, AdminUploadResponse } from '../types'

export function uploadAdminFile(data: AdminUploadRequest, signal?: AbortSignal) {
  return adminHttp.post<AdminUploadResponse>('/util/upload_file', data, {
    signal,
    skipErrorHandler: true,
  })
}
```

- [ ] **Step 6: Create admin auth store**

Create `web/frontend/src/admin/stores/adminAuth.ts`:

```ts
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { loginAdmin, renewAdminAccessToken } from '../api/adminAuthApi'
import type { AdminLoginRequest, AdminTokens, AdminUser } from '../types'

const STORAGE_KEYS = {
  ADMIN: 'nostalgia_admin_user',
  TOKEN: 'nostalgia_admin_access_token',
  TOKEN_EXPIRES: 'nostalgia_admin_access_token_expires_at',
  REFRESH_TOKEN: 'nostalgia_admin_refresh_token',
  REFRESH_TOKEN_EXPIRES: 'nostalgia_admin_refresh_token_expires_at',
} as const

function readJson<T>(key: string): T | null {
  const raw = window.localStorage.getItem(key)
  if (!raw) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    window.localStorage.removeItem(key)
    return null
  }
}

export const useAdminAuthStore = defineStore('adminAuth', () => {
  const admin = ref<AdminUser | null>(readJson<AdminUser>(STORAGE_KEYS.ADMIN))
  const token = ref(window.localStorage.getItem(STORAGE_KEYS.TOKEN) || '')
  const tokenExpiresAt = ref(window.localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRES) || '')
  const refreshToken = ref(window.localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN) || '')
  const refreshTokenExpiresAt = ref(
    window.localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES) || '',
  )

  const isAuthenticated = computed(() => {
    if (!token.value || !tokenExpiresAt.value) return false
    return Date.now() < new Date(tokenExpiresAt.value).getTime()
  })

  const isRefreshTokenValid = computed(() => {
    if (!refreshToken.value || !refreshTokenExpiresAt.value) return false
    return Date.now() < new Date(refreshTokenExpiresAt.value).getTime()
  })

  const setTokens = (tokens: AdminTokens) => {
    token.value = tokens.access_token
    tokenExpiresAt.value = tokens.access_token_expires_at
    refreshToken.value = tokens.refresh_token
    refreshTokenExpiresAt.value = tokens.refresh_token_expires_at
    window.localStorage.setItem(STORAGE_KEYS.TOKEN, tokens.access_token)
    window.localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRES, tokens.access_token_expires_at)
    window.localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, tokens.refresh_token)
    window.localStorage.setItem(
      STORAGE_KEYS.REFRESH_TOKEN_EXPIRES,
      tokens.refresh_token_expires_at,
    )
  }

  const setAdmin = (nextAdmin: AdminUser) => {
    admin.value = nextAdmin
    window.localStorage.setItem(STORAGE_KEYS.ADMIN, JSON.stringify(nextAdmin))
  }

  const clear = () => {
    admin.value = null
    token.value = ''
    tokenExpiresAt.value = ''
    refreshToken.value = ''
    refreshTokenExpiresAt.value = ''
    Object.values(STORAGE_KEYS).forEach((key) => window.localStorage.removeItem(key))
  }

  const login = async (credentials: AdminLoginRequest) => {
    const response = await loginAdmin(credentials)
    const data = response.data
    setTokens(data)
    setAdmin(data.admin)
    return data
  }

  const refreshAccessToken = async () => {
    if (!isRefreshTokenValid.value) {
      clear()
      throw new Error('管理员登录已过期')
    }
    const response = await renewAdminAccessToken(refreshToken.value)
    token.value = response.data.access_token
    tokenExpiresAt.value = response.data.access_token_expires_at
    window.localStorage.setItem(STORAGE_KEYS.TOKEN, response.data.access_token)
    window.localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRES, response.data.access_token_expires_at)
    return response.data.access_token
  }

  const logout = () => {
    clear()
  }

  return {
    admin,
    token,
    tokenExpiresAt,
    refreshToken,
    refreshTokenExpiresAt,
    isAuthenticated,
    isRefreshTokenValid,
    login,
    logout,
    refreshAccessToken,
    setTokens,
    clear,
  }
})
```

- [ ] **Step 7: Create admin HTTP client**

Create `web/frontend/src/admin/api/adminHttp.ts`:

```ts
import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { useToast } from '@/composables/useToast'
import { useAdminAuthStore } from '../stores/adminAuth'

declare module 'axios' {
  interface AxiosRequestConfig {
    skipAuth?: boolean
    skipErrorHandler?: boolean
  }
}

class AdminHttpClient {
  private instance = axios.create({
    baseURL: '/v1',
    timeout: 10000,
  })

  private refreshPromise: Promise<string> | null = null

  constructor() {
    this.instance.interceptors.request.use(this.handleRequest.bind(this))
    this.instance.interceptors.response.use(
      this.handleResponse.bind(this),
      this.handleResponseError.bind(this),
    )
  }

  private async handleRequest(config: InternalAxiosRequestConfig) {
    if (config.skipAuth) return config
    const authStore = useAdminAuthStore()
    const token = authStore.isAuthenticated
      ? authStore.token
      : await this.refreshTokenIfPossible()
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  }

  private handleResponse(response: AxiosResponse) {
    return response
  }

  private async handleResponseError(error: any) {
    const { config, response } = error
    if (response?.status === 401 && config && !config.skipAuth && !config._adminRetry) {
      try {
        const token = await this.refreshTokenIfPossible()
        config._adminRetry = true
        config.headers.Authorization = `Bearer ${token}`
        return this.instance.request(config)
      } catch {
        this.redirectToLogin()
      }
    }

    if (!config?.skipErrorHandler) {
      const toast = useToast()
      toast.add({
        severity: 'error',
        summary: '请求失败',
        detail: response?.data?.error || error.message || '管理员接口请求失败',
      })
    }
    return Promise.reject(error)
  }

  private async refreshTokenIfPossible() {
    if (this.refreshPromise) return this.refreshPromise
    const authStore = useAdminAuthStore()
    this.refreshPromise = authStore.refreshAccessToken()
    try {
      return await this.refreshPromise
    } finally {
      this.refreshPromise = null
    }
  }

  private redirectToLogin() {
    const authStore = useAdminAuthStore()
    authStore.clear()
    const redirect = window.location.pathname + window.location.search
    window.location.href = `/admin/login?redirect=${encodeURIComponent(redirect)}`
  }

  get<T = any>(url: string, config?: any) {
    return this.instance.get<T>(url, config)
  }

  post<T = any>(url: string, data?: any, config?: any) {
    return this.instance.post<T>(url, data, config)
  }

  patch<T = any>(url: string, data?: any, config?: any) {
    return this.instance.patch<T>(url, data, config)
  }

  delete<T = any>(url: string, config?: any) {
    return this.instance.delete<T>(url, config)
  }
}

export default new AdminHttpClient()
```

- [ ] **Step 8: Run frontend type-check and build**

Run:

```bash
cd web/frontend
bun run type-check
bun run build
```

Expected: both commands exit with code 0.

- [ ] **Step 9: Commit Task 1**

Run:

```bash
git add web/frontend/vite.config.ts web/frontend/src/admin
git commit -m "feat(frontend): add admin api client and auth store"
```

## Task 2: Admin Routes, Guard, Layout, and Login

**Files:**
- Create: `web/frontend/src/views/admin/AdminLayout.vue`
- Create: `web/frontend/src/views/admin/AdminLoginView.vue`
- Modify: `web/frontend/src/router/index.ts`
- Modify: `web/frontend/src/App.vue`

- [ ] **Step 1: Add admin route imports and guard**

Modify `web/frontend/src/router/index.ts`:

```ts
import { useAdminAuthStore } from '@/admin/stores/adminAuth'
```

Add these routes before the catch-all route:

```ts
{
  path: '/admin/login',
  name: 'adminLogin',
  component: () => import('@/views/admin/AdminLoginView.vue'),
  meta: { hideNavbar: true, hideFooter: true },
},
{
  path: '/admin',
  component: () => import('@/views/admin/AdminLayout.vue'),
  meta: { hideNavbar: true, hideFooter: true, requiresAdmin: true },
  children: [
    { path: '', redirect: { name: 'adminArticles' } },
    {
      path: 'articles',
      name: 'adminArticles',
      component: () => import('@/views/admin/AdminArticleListView.vue'),
    },
    {
      path: 'articles/new',
      name: 'adminArticleNew',
      component: () => import('@/views/admin/AdminArticleEditorView.vue'),
    },
    {
      path: 'articles/:id/edit',
      name: 'adminArticleEdit',
      component: () => import('@/views/admin/AdminArticleEditorView.vue'),
      props: true,
    },
    {
      path: 'categories',
      name: 'adminCategories',
      component: () => import('@/views/admin/AdminCategoryView.vue'),
    },
  ],
},
```

Add this guard after router creation:

```ts
router.beforeEach((to) => {
  if (!to.meta.requiresAdmin) return true
  const adminAuth = useAdminAuthStore()
  if (adminAuth.isAuthenticated) return true
  return {
    name: 'adminLogin',
    query: { redirect: to.fullPath },
  }
})
```

- [ ] **Step 2: Hide footer for admin routes**

Modify `web/frontend/src/App.vue`:

```vue
<footer v-if="!route.meta.hideFooter">
  <FooterView />
</footer>
```

- [ ] **Step 3: Create admin layout**

Create `web/frontend/src/views/admin/AdminLayout.vue` with navigation links for articles and categories, a theme switcher, a site link, and logout. Use `RouterView`, `RouterLink`, `useRoute`, `useRouter`, `ThemeSwitcher`, `AppButton`, and lucide icons `BookOpen`, `FolderTree`, `ExternalLink`, `LogOut`, `Menu`, `X`.

Required behaviors:

- Desktop: sidebar visible at `lg` width.
- Mobile: top bar with menu toggle.
- Logout clears admin auth and routes to `/admin/login`.
- Active link uses `bg-accent/10 text-accent`.

- [ ] **Step 4: Create admin login view**

Create `web/frontend/src/views/admin/AdminLoginView.vue` with:

- `username` and `password` fields.
- Submit calls `useAdminAuthStore().login`.
- Success redirects to `route.query.redirect` or `/admin/articles`.
- Failure shows toast: summary `登录失败`, detail from backend error or `管理员账号或密码不正确`.
- The page uses current logo `/logo.svg`, `ArchivePanel`, `AppInput`, and `AppButton`.

- [ ] **Step 5: Create temporary admin page stubs**

Create these temporary files so routes type-check before real pages are implemented:

```vue
<!-- web/frontend/src/views/admin/AdminArticleListView.vue -->
<template><div class="p-6">文章管理</div></template>
```

```vue
<!-- web/frontend/src/views/admin/AdminArticleEditorView.vue -->
<template><div class="p-6">文章编辑</div></template>
```

```vue
<!-- web/frontend/src/views/admin/AdminCategoryView.vue -->
<template><div class="p-6">分类管理</div></template>
```

- [ ] **Step 6: Run frontend verification**

Run:

```bash
cd web/frontend
bun run type-check
bun run build
```

Expected: both commands exit with code 0.

- [ ] **Step 7: Commit Task 2**

Run:

```bash
git add web/frontend/src/router/index.ts web/frontend/src/App.vue web/frontend/src/views/admin
git commit -m "feat(frontend): add admin layout and route guard"
```

## Task 3: Admin Article List

**Files:**
- Modify: `web/frontend/src/views/admin/AdminArticleListView.vue`

- [ ] **Step 1: Replace article list stub with state and API loading**

Implement:

- `articles = ref<AdminArticle[]>([])`
- `loading = ref(false)`
- `query = reactive({ title: '', page: 1, limit: 12 })`
- `total = ref(0)`
- `fetchArticles()` calls `listAdminArticles(query)` and stores `response.data.articles`.
- Convert `response.data.count` with `Number(response.data.count || 0)`.

- [ ] **Step 2: Add list UI**

Render:

- Header with title `文章管理`, search input, and `写新文章` button.
- A responsive list of article rows using glass/surface styling.
- Each row shows title, summary, category, views, likes, updated date, and publish badge.
- Actions: edit, publish/draft toggle, delete.

Use existing components:

```ts
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { useToast } from '@/composables/useToast'
```

- [ ] **Step 3: Implement article actions**

Required methods:

- `createArticle()` routes to `{ name: 'adminArticleNew' }`.
- `editArticle(id)` routes to `{ name: 'adminArticleEdit', params: { id } }`.
- `togglePublish(article)` calls `updateAdminArticle({ id: article.id, is_publish: !article.is_publish })`, then updates local item.
- `askDelete(article)` opens `ConfirmDialog`.
- `confirmDelete()` calls `deleteAdminArticle(selected.id)`, removes local item, and shows success toast.

- [ ] **Step 4: Add empty, loading, and paging states**

Render:

- Loading text `正在读取文章`.
- Empty text `还没有匹配的文章`.
- Previous/next buttons using `query.page`, `query.limit`, and `total`.

- [ ] **Step 5: Run frontend verification**

Run:

```bash
cd web/frontend
bun run type-check
bun run build
```

Expected: both commands exit with code 0.

- [ ] **Step 6: Commit Task 3**

Run:

```bash
git add web/frontend/src/views/admin/AdminArticleListView.vue
git commit -m "feat(frontend): add admin article management"
```

## Task 4: Admin Article Editor

**Files:**
- Create: `web/frontend/src/admin/editor/adminEditorConfig.ts`
- Create: `web/frontend/src/admin/editor/adminUploadAdapter.ts`
- Modify: `web/frontend/src/views/admin/AdminArticleEditorView.vue`
- Modify: `web/frontend/src/assets/content.css`

- [ ] **Step 1: Create editor config**

Create `web/frontend/src/admin/editor/adminEditorConfig.ts` by adapting `web/backend/src/config/editorConfig.ts`:

- Keep toolbar items: undo, redo, heading, bold, italic, underline, strikethrough, removeFormat, bulletedList, numberedList, todoList, outdent, indent, link, insertImage, insertTable, blockQuote, codeBlock, horizontalLine, alignment, sourceEditing.
- Keep language `zh-cn`.
- Keep code block languages for plaintext, go, python, javascript, typescript, java, c, cpp, sql, json, bash, html, css.
- Keep image and table toolbars.
- Keep CKEditor empty-state text `这一刻的想法……`.

- [ ] **Step 2: Create upload adapter**

Create `web/frontend/src/admin/editor/adminUploadAdapter.ts`:

```ts
import { uploadAdminFile } from '../api/adminUploadApi'

export default class AdminUploadAdapter {
  private abortController = new AbortController()

  constructor(
    private loader: any,
    private articleId: string,
    private type: 'content' | 'cover' = 'content',
  ) {}

  async upload() {
    const file = (await this.loader.file) as File | null
    if (!file) throw new Error('请选择要上传的图片')
    if (!file.type.startsWith('image/')) throw new Error('只能上传图片文件')
    this.loader.uploadTotal = file.size
    const content = await this.fileToBase64(file)
    const response = await uploadAdminFile(
      { article_id: this.articleId, content, type: this.type },
      this.abortController.signal,
    )
    return { default: response.data.url, url: response.data.url }
  }

  abort() {
    this.abortController.abort()
  }

  private fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.readAsDataURL(file)
      reader.onload = () => {
        this.loader.uploaded = file.size
        resolve(String(reader.result).split(',')[1] || '')
      }
      reader.onprogress = (event) => {
        if (event.lengthComputable) this.loader.uploaded = event.loaded
      }
      reader.onerror = () => reject(new Error('读取图片失败'))
    })
  }
}
```

- [ ] **Step 3: Implement editor data flow**

Replace the editor stub with:

- Route param `id`.
- If route name is `adminArticleNew`, call `createAdminArticle({ title: '无标题文章', is_publish: false })`, then `router.replace({ name: 'adminArticleEdit', params: { id: article.id } })`.
- If editing existing article, call `getAdminArticle(id, true)`.
- Load categories with `listAllAdminCategories()`.
- Keep `article = ref<Partial<AdminArticle>>({ id: '', title: '', summary: '', content: '', category_id: undefined, is_publish: false, cover: '', slug: '', check_outdated: false })`.
- Keep `editorData = ref('')`.
- Keep `isSaving`, `isLayoutReady`, `charCount`, `wordCount`, `saveStatus`, `hasDraft`.

- [ ] **Step 4: Implement draft cache and leave guard**

Use key:

```ts
const draftKey = computed(() => `nostalgia_admin_article_draft:${article.value.id || 'new'}`)
```

Required behavior:

- Watch `editorData` and `article` deeply after initial load.
- Save changed content to `sessionStorage`.
- Restore draft when draft article id matches the current article id.
- `onBeforeRouteLeave` asks for confirmation with `window.confirm('文章还有未保存的修改，确认离开吗？')` when dirty.

- [ ] **Step 5: Implement save and cover upload**

Save:

```ts
await updateAdminArticle({ ...article.value, content: editorData.value })
```

Cover upload:

- Validate image file.
- Convert to base64.
- Call `uploadAdminFile({ article_id: article.value.id!, content, type: 'cover' })`.
- Set `article.value.cover = response.data.url + '?t=' + Date.now()`.

- [ ] **Step 6: Implement editor UI**

UI requirements:

- Top sticky editor bar with title input, save button, back button, publish state badge, save status.
- Main content area with CKEditor wrapped in a class that applies article reading rhythm.
- Right metadata panel with summary textarea, slug input, category select, publish checkbox, outdated checkbox, and cover controls.
- Mobile layout stacks metadata under editor.

- [ ] **Step 7: Add CKEditor style bridge**

Modify `web/frontend/src/assets/content.css` with admin editor bridge selectors:

```css
.admin-editor-content .ck-content {
  min-height: 58vh;
}

.admin-editor-content .ck-content,
.admin-editor-content .ck-content.ck-editor__editable {
  color: rgb(var(--color-foreground));
}

.admin-editor-content .ck-content h1,
.admin-editor-content .ck-content h2,
.admin-editor-content .ck-content h3,
.admin-editor-content .ck-content h4,
.admin-editor-content .ck-content h5,
.admin-editor-content .ck-content h6,
.admin-editor-content .ck-content p,
.admin-editor-content .ck-content li,
.admin-editor-content .ck-content blockquote,
.admin-editor-content .ck-content pre,
.admin-editor-content .ck-content table,
.admin-editor-content .ck-content figure {
  max-width: 100%;
}
```

- [ ] **Step 8: Run frontend verification**

Run:

```bash
cd web/frontend
bun run type-check
bun run build
```

Expected: both commands exit with code 0.

- [ ] **Step 9: Commit Task 4**

Run:

```bash
git add web/frontend/src/admin/editor web/frontend/src/views/admin/AdminArticleEditorView.vue web/frontend/src/assets/content.css
git commit -m "feat(frontend): add admin article editor"
```

## Task 5: Admin Category Management

**Files:**
- Modify: `web/frontend/src/views/admin/AdminCategoryView.vue`

- [ ] **Step 1: Replace category stub with data loading**

Implement:

- `categories = ref<AdminCategory[]>([])`
- `loading = ref(false)`
- `editingId = ref<number | null>(null)`
- `draftName = ref('')`
- `newCategoryName = ref('')`
- `fetchCategories()` calls `listAllAdminCategories()`.

- [ ] **Step 2: Implement category create, rename, and delete**

Required methods:

- `createCategory()` trims `newCategoryName`, calls `createAdminCategory`, resets input, refetches categories, and shows success toast.
- `startEdit(category)` sets `editingId` and `draftName`.
- `saveEdit(category)` calls `updateAdminCategory({ id: category.id, name: draftName.trim() })`, clears editing state, refetches, and shows success toast.
- `askDelete(category)` opens `ConfirmDialog`.
- `confirmDelete()` calls `deleteAdminCategory(selected.id)`, refetches, and shows success toast.

- [ ] **Step 3: Implement category UI**

Render:

- Header `分类管理`.
- Inline create form.
- Category rows with name, article count, created time, and edit/delete actions.
- Inline rename field when editing.
- Empty text `还没有分类`.

- [ ] **Step 4: Run frontend verification**

Run:

```bash
cd web/frontend
bun run type-check
bun run build
```

Expected: both commands exit with code 0.

- [ ] **Step 5: Commit Task 5**

Run:

```bash
git add web/frontend/src/views/admin/AdminCategoryView.vue
git commit -m "feat(frontend): add admin category management"
```

## Task 6: Legacy Admin Cleanup and Documentation

**Files:**
- Delete: `web/backend/`
- Modify: `AGENTS.md`
- Inspect and modify if references exist: `README.md`, `.github/workflows/*`, deployment docs, Docker/Caddy/Nginx config files

- [ ] **Step 1: Search for legacy backend frontend references**

Run:

```bash
rg -n "web/backend|backend admin|Geeker|Element Plus|build:dev|npm run build:dev|VITE_API_URL|/backend" README.md AGENTS.md .github docker* deploy* web -S
```

Expected: command prints all remaining references that need review.

- [ ] **Step 2: Delete legacy admin frontend directory**

Run:

```bash
rm -rf web/backend
```

- [ ] **Step 3: Update `AGENTS.md`**

Edit `AGENTS.md`:

- Change project description from two Vue frontend apps to one Vue frontend app with public and `/admin` routes.
- Remove backend frontend commands.
- Keep `cd web/frontend && bun run type-check && bun run build`.
- Document that `/v1` is the admin gRPC-Gateway API surface.

- [ ] **Step 4: Update deployment references found in Step 1**

For every reference from Step 1 that points to building or serving `web/backend`, remove it or redirect it to the unified `web/frontend` app. Keep backend service references for Go, Gin, gRPC, and gRPC-Gateway.

- [ ] **Step 5: Run full frontend verification**

Run:

```bash
cd web/frontend
bun run type-check
bun run build
```

Expected: both commands exit with code 0.

- [ ] **Step 6: Run git status review**

Run:

```bash
git status --short
```

Expected: removed files are under `web/backend/`; modified docs/config files match the references found in Step 1.

- [ ] **Step 7: Commit Task 6**

Run:

```bash
git add -A web/backend AGENTS.md README.md .github docker-compose.yaml docker-compose.dev.yaml Caddyfile nginx.conf
git status --short
git commit -m "chore(frontend): remove legacy backend admin app"
```

If some listed files do not exist, omit them from `git add` and stage only files that exist.

## Final Verification

- [ ] **Step 1: Run frontend type-check and build**

Run:

```bash
cd web/frontend
bun run type-check
bun run build
```

Expected: both commands exit with code 0.

- [ ] **Step 2: Run backend tests if any backend or deployment routing changed**

Run:

```bash
make test
```

Expected: command exits with code 0. If local database or Redis dependencies prevent this command from running, record the exact error in the final report.

- [ ] **Step 3: Run manual smoke test with local services**

With API services running:

```bash
cd web/frontend
bun run dev
```

Verify in browser:

- Public `/` loads.
- Public article detail loads and keeps reading styles.
- `/admin/login` loads without public nav/footer.
- Admin login succeeds.
- `/admin/articles` lists articles.
- New article draft creates and opens editor.
- CKEditor image upload calls `/v1/util/upload_file`.
- Saving article updates public rendering.
- `/admin/categories` can create, rename, and delete a category.
- Light/dark/system theme mode persists across public and admin routes.

- [ ] **Step 4: Push branch and create PR**

Run:

```bash
git status --short
git push -u origin feature/unified-admin-frontend-migration
gh pr create --base master --head feature/unified-admin-frontend-migration --title "feat: migrate admin frontend into public app" --body-file -
```

Use a PR body with Summary, Test Plan, and Notes sections listing the verification commands and any manual smoke-test results.
