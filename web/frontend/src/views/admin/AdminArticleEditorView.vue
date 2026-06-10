<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ArrowLeft, ImagePlus, Save, Trash2 } from '@lucide/vue'
import { Ckeditor } from '@ckeditor/ckeditor5-vue'
import { ClassicEditor, type EditorConfig } from 'ckeditor5'
import 'ckeditor5/ckeditor5.css'
import 'ckeditor5/ckeditor5-content.css'
import type { AdminArticle, AdminCategory } from '@/admin/types'
import {
  createAdminArticle,
  getAdminArticle,
  updateAdminArticle
} from '@/admin/api/adminArticleApi'
import { listAllAdminCategories } from '@/admin/api/adminCategoryApi'
import { uploadAdminFile } from '@/admin/api/adminUploadApi'
import { adminEditorConfig } from '@/admin/editor/adminEditorConfig'
import AdminUploadAdapter from '@/admin/editor/adminUploadAdapter'
import {
  ADMIN_IMAGE_ACCEPT,
  getAdminUploadErrorMessage,
  validateAdminImageFile
} from '@/admin/editor/uploadPolicy'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { useToast } from '@/composables/useToast'

interface DraftPayload {
  id: string
  content: string
  article: Partial<AdminArticle>
}

const route = useRoute()
const router = useRouter()
const toast = useToast()

const defaultArticle = (): Partial<AdminArticle> => ({
  id: '',
  title: '',
  summary: '',
  content: '',
  category_id: undefined,
  is_publish: false,
  cover: '',
  slug: '',
  check_outdated: false
})

const article = ref<Partial<AdminArticle>>(defaultArticle())
const categories = ref<AdminCategory[]>([])
const editorData = ref('')
const isSaving = ref(false)
const isCoverUploading = ref(false)
const isLayoutReady = ref(false)
const isTrackingChanges = ref(false)
const charCount = ref(0)
const wordCount = ref(0)
const saveStatus = ref<'saved' | 'dirty' | 'saving' | 'error'>('saved')
const hasDraft = ref(false)
const lastSavedAt = ref('')
const loadError = ref('')
const coverInput = ref<HTMLInputElement | null>(null)

let initialContent = ''
let initialArticle = ''

const articleTitle = computed({
  get: () => article.value.title || '',
  set: (value: string) => {
    article.value.title = value
  }
})

const articleSummary = computed({
  get: () => article.value.summary || '',
  set: (value: string) => {
    article.value.summary = value
  }
})

const articleSlug = computed({
  get: () => article.value.slug || '',
  set: (value: string) => {
    article.value.slug = value
  }
})

const isPublished = computed({
  get: () => Boolean(article.value.is_publish),
  set: (value: boolean) => {
    article.value.is_publish = value
  }
})

const shouldCheckOutdated = computed({
  get: () => Boolean(article.value.check_outdated),
  set: (value: boolean) => {
    article.value.check_outdated = value
  }
})

const draftKey = computed(() => `nostalgia_admin_article_draft:${article.value.id || 'new'}`)

const saveStatusText = computed(() => {
  if (saveStatus.value === 'saving') return '保存中...'
  if (saveStatus.value === 'error') return '保存失败，修改仍已缓存'
  if (hasUnsavedChanges()) return hasDraft.value ? '有未保存草稿' : '有未保存修改'
  if (lastSavedAt.value) return `已保存 ${lastSavedAt.value}`
  return '已保存'
})

const saveStatusTone = computed<'neutral' | 'accent' | 'danger' | 'warning'>(() => {
  if (saveStatus.value === 'error') return 'danger'
  if (saveStatus.value === 'saving' || hasUnsavedChanges()) return 'warning'
  return 'accent'
})

const canSave = computed(() => {
  return Boolean(article.value.id && isLayoutReady.value && !isSaving.value)
})

const coverPreview = computed(() => article.value.cover || '')

const snapshotArticle = () =>
  JSON.stringify({
    title: article.value.title || '',
    summary: article.value.summary || '',
    category_id: article.value.category_id ?? null,
    is_publish: Boolean(article.value.is_publish),
    cover: article.value.cover || '',
    slug: article.value.slug || '',
    check_outdated: Boolean(article.value.check_outdated)
  })

const hasUnsavedChanges = () => {
  return (
    isTrackingChanges.value &&
    (editorData.value !== initialContent || snapshotArticle() !== initialArticle)
  )
}

const clearDraft = () => {
  sessionStorage.removeItem(draftKey.value)
  hasDraft.value = false
}

const restoreDraft = () => {
  const raw = sessionStorage.getItem(draftKey.value)
  if (!raw || !article.value.id) return

  try {
    const draft = JSON.parse(raw) as DraftPayload

    if (draft.id !== article.value.id) return

    const nextArticle = { ...article.value, ...draft.article, id: article.value.id }
    const hasDifferentContent =
      draft.content !== editorData.value ||
      JSON.stringify(nextArticle) !== JSON.stringify(article.value)

    if (!hasDifferentContent) {
      clearDraft()
      return
    }

    article.value = nextArticle
    editorData.value = draft.content || ''
    hasDraft.value = true
    saveStatus.value = 'dirty'
    toast.add({
      severity: 'info',
      summary: '已恢复草稿',
      detail: '上次未保存的编辑内容已恢复',
      life: 2600
    })
  } catch {
    clearDraft()
  }
}

const trackDraft = () => {
  if (!isTrackingChanges.value || !article.value.id) return

  if (!hasUnsavedChanges()) {
    if (saveStatus.value !== 'saving') saveStatus.value = 'saved'
    clearDraft()
    return
  }

  sessionStorage.setItem(
    draftKey.value,
    JSON.stringify({
      id: article.value.id,
      content: editorData.value,
      article: { ...article.value, content: editorData.value }
    })
  )

  hasDraft.value = true
  if (saveStatus.value !== 'saving' && saveStatus.value !== 'error') {
    saveStatus.value = 'dirty'
  }
}

watch([editorData, article], trackDraft, { deep: true })

const resetInitialState = () => {
  initialContent = editorData.value
  initialArticle = snapshotArticle()
  saveStatus.value = 'saved'
  lastSavedAt.value = ''
  hasDraft.value = false
}

const applyArticle = (nextArticle: AdminArticle) => {
  article.value = { ...defaultArticle(), ...nextArticle }
  editorData.value = nextArticle.content || ''
}

const fetchCategories = async () => {
  try {
    const response = await listAllAdminCategories()
    categories.value = response.data.categories ?? []
  } catch {
    categories.value = []
  }
}

const getRouteArticleId = () => {
  return typeof route.params.id === 'string' ? route.params.id : ''
}

const loadEditor = async () => {
  loadError.value = ''
  isLayoutReady.value = false
  isTrackingChanges.value = false
  article.value = defaultArticle()
  editorData.value = ''

  try {
    await fetchCategories()

    if (route.name === 'adminArticleNew') {
      const response = await createAdminArticle({ title: '无标题文章', is_publish: false })
      applyArticle(response.data.article)
      await router.replace({ name: 'adminArticleEdit', params: { id: response.data.article.id } })
    } else {
      const articleId = getRouteArticleId()
      if (!articleId) throw new Error('缺少文章 ID')
      const response = await getAdminArticle(articleId, true)
      applyArticle(response.data.article)
    }

    resetInitialState()
    restoreDraft()
    isTrackingChanges.value = true
    saveStatus.value = hasUnsavedChanges() ? 'dirty' : 'saved'
    isLayoutReady.value = true
  } catch {
    loadError.value = '文章加载失败'
  }
}

function WordCountPlugin(editor: any) {
  const wordCountPlugin = editor.plugins.get('WordCount')
  wordCountPlugin.on('update', (_event: unknown, stats: { characters: number; words: number }) => {
    charCount.value = stats.characters
    wordCount.value = stats.words
  })
}

function AdminUploadAdapterPlugin(editor: any) {
  editor.plugins.get('FileRepository').createUploadAdapter = (loader: any) => {
    return new AdminUploadAdapter(loader, String(article.value.id), 'content')
  }
}

const editorConfig = computed<EditorConfig>(() => ({
  ...adminEditorConfig,
  extraPlugins: [AdminUploadAdapterPlugin, WordCountPlugin]
}))

const onEditorReady = (editorInstance: ClassicEditor) => {
  editorInstance.ui.view.editable.element?.classList.add('reading-prose')
}

const formatTime = () => {
  const date = new Date()
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${hours}:${minutes}:${seconds}`
}

const saveArticle = async () => {
  if (!canSave.value) return

  const title = articleTitle.value.trim()
  if (!title) {
    toast.add({
      severity: 'warning',
      summary: '标题不能为空',
      detail: '请先给文章起一个标题',
      life: 2600
    })
    return
  }

  isSaving.value = true
  saveStatus.value = 'saving'

  try {
    await updateAdminArticle({
      id: String(article.value.id),
      title,
      summary: articleSummary.value,
      content: editorData.value,
      category_id: article.value.category_id,
      is_publish: isPublished.value,
      cover: article.value.cover || '',
      slug: articleSlug.value,
      check_outdated: shouldCheckOutdated.value
    })

    article.value.content = editorData.value
    initialContent = editorData.value
    initialArticle = snapshotArticle()
    lastSavedAt.value = formatTime()
    saveStatus.value = 'saved'
    clearDraft()
    toast.add({
      severity: 'success',
      summary: '保存成功',
      detail: title,
      life: 2400
    })
  } catch (error) {
    saveStatus.value = 'error'
    toast.add({
      severity: 'error',
      summary: '保存失败',
      detail: getAdminUploadErrorMessage(error, '修改已保存在本地草稿，请稍后重试'),
      life: 3200
    })
  } finally {
    isSaving.value = false
  }
}

const fileToBase64 = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(String(reader.result).split(',')[1] || '')
    reader.onerror = () => reject(new Error('读取图片失败'))
  })
}

const openCoverPicker = () => {
  coverInput.value?.click()
}

const handleCoverInput = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''

  const validationMessage = validateAdminImageFile(file)
  if (validationMessage) {
    toast.add({
      severity: 'warning',
      summary: '封面上传失败',
      detail: validationMessage,
      life: 2600
    })
    return
  }

  if (!article.value.id) return

  isCoverUploading.value = true

  try {
    const content = await fileToBase64(file!)
    const response = await uploadAdminFile({
      article_id: String(article.value.id),
      content,
      type: 'cover'
    })
    article.value.cover = `${response.data.url}?t=${Date.now()}`
    toast.add({
      severity: 'success',
      summary: '封面已更新',
      detail: '保存文章后封面变更会正式生效',
      life: 2600
    })
  } catch (error: any) {
    toast.add({
      severity: 'error',
      summary: '封面上传失败',
      detail: getAdminUploadErrorMessage(error),
      life: 3000
    })
  } finally {
    isCoverUploading.value = false
  }
}

const removeCover = () => {
  article.value.cover = ''
}

const updateCategory = (event: Event) => {
  const value = (event.target as HTMLSelectElement).value
  article.value.category_id = value || undefined
}

const goBack = () => {
  void router.push({ name: 'adminArticles' })
}

const handleSaveShortcut = (event: KeyboardEvent) => {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
    event.preventDefault()
    void saveArticle()
  }
}

const handleBeforeUnload = (event: BeforeUnloadEvent) => {
  if (!hasUnsavedChanges() || isSaving.value) return
  event.preventDefault()
  event.returnValue = ''
}

onBeforeRouteLeave((_to, _from, next) => {
  if (!hasUnsavedChanges() || isSaving.value) {
    next()
    return
  }

  if (window.confirm('文章还有未保存的修改，确认离开吗？')) {
    next()
  } else {
    next(false)
  }
})

onMounted(() => {
  void loadEditor()
  window.addEventListener('keydown', handleSaveShortcut)
  window.addEventListener('beforeunload', handleBeforeUnload)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleSaveShortcut)
  window.removeEventListener('beforeunload', handleBeforeUnload)
})
</script>

<template>
  <main class="space-y-5">
    <header
      class="sticky top-0 z-30 -mx-4 border-b border-border/70 bg-background/95 px-4 py-3 backdrop-blur sm:-mx-6 sm:px-6 lg:-mx-8 lg:px-8"
    >
      <div
        class="mx-auto flex max-w-7xl flex-col gap-3 xl:flex-row xl:items-center xl:justify-between"
      >
        <div class="flex min-w-0 flex-1 flex-col gap-3 sm:flex-row sm:items-center">
          <AppButton variant="ghost" size="sm" class="w-max" @click="goBack">
            <ArrowLeft class="size-4" aria-hidden="true" />
            返回
          </AppButton>

          <label class="min-w-0 flex-1">
            <span class="sr-only">文章标题</span>
            <AppInput
              v-model="articleTitle"
              placeholder="无标题文章"
              class="h-12 rounded-archive text-base font-black sm:text-lg"
              :disabled="!isLayoutReady"
            />
          </label>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <AppBadge :tone="isPublished ? 'accent' : 'neutral'">
            {{ isPublished ? '已发布' : '草稿' }}
          </AppBadge>
          <AppBadge :tone="saveStatusTone">{{ saveStatusText }}</AppBadge>
          <AppBadge tone="neutral" class="tabular-nums">
            {{ charCount }} 字符 / {{ wordCount }} 词
          </AppBadge>
          <AppButton :disabled="!canSave" @click="saveArticle">
            <Save class="size-4" aria-hidden="true" />
            {{ isSaving ? '保存中...' : '保存文章' }}
          </AppButton>
        </div>
      </div>
    </header>

    <section v-if="loadError" class="archive-surface rounded-archive p-8 text-center">
      <p class="m-0 text-lg font-black text-foreground">{{ loadError }}</p>
      <p class="m-0 mt-2 text-sm text-muted-foreground">请返回列表后重新进入编辑器。</p>
      <AppButton class="mt-5" variant="secondary" @click="goBack">返回文章列表</AppButton>
    </section>

    <section v-else-if="!isLayoutReady" class="archive-surface rounded-archive p-8 text-center">
      <p class="m-0 text-sm font-semibold text-muted-foreground">正在准备编辑器</p>
    </section>

    <section v-else class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_22rem]">
      <div class="min-w-0 space-y-3">
        <div class="archive-surface admin-editor-content overflow-hidden rounded-archive">
          <Ckeditor
            v-model="editorData"
            :editor="ClassicEditor"
            :config="editorConfig"
            @ready="onEditorReady"
          />
        </div>
      </div>

      <aside class="archive-surface h-max rounded-archive p-4 xl:sticky xl:top-24">
        <h2 class="m-0 text-base font-black text-foreground">文章设置</h2>
        <div class="mt-4 space-y-4">
          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">摘要</span>
            <textarea
              v-model="articleSummary"
              rows="5"
              class="w-full resize-y rounded-archive border border-border bg-surface px-4 py-3 text-sm leading-6 text-foreground outline-none transition-colors placeholder:text-muted-foreground/80 focus:border-accent focus:ring-2 focus:ring-accent/20"
              placeholder="简短说明这篇文章解决什么问题"
            />
          </label>

          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">短链接 Slug</span>
            <AppInput v-model="articleSlug" placeholder="custom-url-slug" />
          </label>

          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">分类</span>
            <select
              :value="article.category_id ?? ''"
              class="h-11 w-full rounded-full border border-border bg-surface px-4 text-sm text-foreground outline-none transition-colors focus:border-accent focus:ring-2 focus:ring-accent/20"
              @change="updateCategory"
            >
              <option value="">不设置分类</option>
              <option
                v-for="category in categories"
                :key="category.id"
                :value="String(category.id)"
              >
                {{ category.name }}
              </option>
            </select>
          </label>

          <div class="space-y-3 border-y border-border/70 py-4">
            <label class="flex items-start gap-3">
              <input
                v-model="isPublished"
                type="checkbox"
                class="mt-1 size-4 rounded border-border accent-[rgb(var(--color-accent))]"
              />
              <span>
                <span class="block text-sm font-bold text-foreground">发布文章</span>
                <span class="block text-xs leading-5 text-muted-foreground"
                  >关闭后文章会保存为草稿。</span
                >
              </span>
            </label>

            <label class="flex items-start gap-3">
              <input
                v-model="shouldCheckOutdated"
                type="checkbox"
                class="mt-1 size-4 rounded border-border accent-[rgb(var(--color-accent))]"
              />
              <span>
                <span class="block text-sm font-bold text-foreground">检查时效</span>
                <span class="block text-xs leading-5 text-muted-foreground"
                  >适合会随时间变化的技术内容。</span
                >
              </span>
            </label>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between gap-3">
              <span class="text-sm font-bold text-foreground">封面图</span>
              <input
                ref="coverInput"
                type="file"
                :accept="ADMIN_IMAGE_ACCEPT"
                class="hidden"
                @change="handleCoverInput"
              />
              <AppButton
                variant="secondary"
                size="sm"
                :disabled="isCoverUploading"
                @click="openCoverPicker"
              >
                <ImagePlus class="size-4" aria-hidden="true" />
                {{ isCoverUploading ? '上传中...' : '上传' }}
              </AppButton>
            </div>

            <div
              v-if="coverPreview"
              class="overflow-hidden rounded-archive border border-border bg-muted"
            >
              <img
                :src="coverPreview"
                :alt="articleTitle || '文章封面'"
                class="aspect-[16/9] w-full object-cover"
              />
              <div class="flex justify-end border-t border-border bg-surface p-2">
                <AppButton
                  variant="ghost"
                  size="sm"
                  class="text-danger hover:text-danger"
                  @click="removeCover"
                >
                  <Trash2 class="size-4" aria-hidden="true" />
                  移除封面
                </AppButton>
              </div>
            </div>

            <div
              v-else
              class="rounded-archive border border-dashed border-border bg-surface-raised p-5 text-center"
            >
              <p class="m-0 text-sm font-semibold text-muted-foreground">还没有封面图</p>
            </div>
          </div>
        </div>
      </aside>
    </section>
  </main>
</template>
