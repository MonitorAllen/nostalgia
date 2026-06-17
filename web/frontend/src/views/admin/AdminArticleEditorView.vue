<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Copy, Eye, ImagePlus, Save, Sparkles, Trash2, Wand2, X } from '@lucide/vue'
import { Ckeditor } from '@ckeditor/ckeditor5-vue'
import { ClassicEditor, type EditorConfig } from 'ckeditor5'
import 'ckeditor5/ckeditor5.css'
import 'ckeditor5/ckeditor5-content.css'
import type {
  AdminAIPolishMode,
  AdminAIPolishSuggestion,
  AdminAIPolishTarget,
  AdminArticle,
  AdminCategory
} from '@/admin/types'
import {
  createAdminArticle,
  getAdminArticle,
  updateAdminArticle
} from '@/admin/api/adminArticleApi'
import { polishAdminText } from '@/admin/api/adminAiApi'
import { listAllAdminCategories } from '@/admin/api/adminCategoryApi'
import { uploadAdminFile } from '@/admin/api/adminUploadApi'
import {
  buildAIPolishRequest,
  buildAIPolishContentPreview,
  createAIPolishSession,
  getAIPolishApplyLabel,
  getAIPolishModeLabel,
  normalizeSelectedText,
  truncateForAIPolish,
  type AIPolishSession
} from '@/admin/ai/polish'
import { adminEditorConfig } from '@/admin/editor/adminEditorConfig'
import AdminUploadAdapter from '@/admin/editor/adminUploadAdapter'
import {
  ADMIN_IMAGE_ACCEPT,
  getAdminUploadErrorMessage,
  validateAdminImageFile
} from '@/admin/editor/uploadPolicy'
import {
  getSuggestionPreviewHtml,
  insertSuggestionContent
} from '@/admin/editor/markdownRichText'
import { htmlToPlainText, writeRichClipboard } from '@/admin/editor/richClipboard'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { useToast } from '@/composables/useToast'
import { sanitizeHtml } from '@/util/sanitizeHtml'

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
const editorInstance = ref<ClassicEditor | null>(null)
const lastEditorSelectionRange = ref<any | null>(null)
const lastEditorSelectionText = ref('')
const lastEditorSelectionHtml = ref('')
const aiPolishMarkerName = ref('')
const isPolishing = ref(false)
const aiDrawerOpen = ref(false)
const aiPolishSession = ref<AIPolishSession | null>(null)
const polishTarget = ref<AdminAIPolishTarget>('content_selection')
const polishMode = ref<AdminAIPolishMode>('improve')
const polishSuggestions = ref<AdminAIPolishSuggestion[]>([])
const selectedSuggestionIndex = ref(-1)
const previewOpen = ref(false)
const previewOverrides = ref<{ title?: string; summary?: string; content?: string } | null>(null)
const savedIsPublished = ref(false)

let initialContent = ''
let initialArticle = ''
let aiPolishMarkerSeed = 0

const AI_POLISH_MARKER_GROUP = 'ai-polish-target'

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
const previewTitle = computed(() => previewOverrides.value?.title ?? articleTitle.value)
const previewSummary = computed(() => previewOverrides.value?.summary ?? articleSummary.value)
const previewContent = computed(() =>
  sanitizeHtml(previewOverrides.value?.content ?? editorData.value ?? '')
)
const adminArticleListLocation = computed(() => ({ name: 'adminArticles', query: route.query }))
const savedPublishTone = computed(() => (savedIsPublished.value ? 'accent' : 'neutral'))
const savedPublishText = computed(() => (savedIsPublished.value ? '已发布' : '草稿'))

const polishPanelTitle = computed(() => {
  if (polishTarget.value === 'title') return '标题候选'
  if (polishTarget.value === 'summary') return '摘要候选'
  return getAIPolishModeLabel(polishMode.value)
})

const canUseAIPolish = computed(() => isLayoutReady.value && !isPolishing.value)

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
  savedIsPublished.value = Boolean(nextArticle.is_publish)
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
      await router.replace({
        name: 'adminArticleEdit',
        params: { id: response.data.article.id },
        query: route.query
      })
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

const installAIPolishMarkerHighlight = (editor: any) => {
  const conversion = editor.conversion.for('editingDowncast')
  if (!conversion?.markerToHighlight) return

  conversion.markerToHighlight({
    model: AI_POLISH_MARKER_GROUP,
    view: {
      classes: ['admin-ai-polish-target']
    },
    converterPriority: 'high'
  })
}

const onEditorReady = (readyEditor: ClassicEditor) => {
  installAIPolishMarkerHighlight(readyEditor)
  readyEditor.ui.view.editable.element?.classList.add('admin-editor-prose', 'ck-content')
  editorInstance.value = readyEditor
  const selection = (readyEditor as any).model.document.selection
  selection.on('change:range', rememberEditorSelection)
}

const getArticleTextExcerpt = () => {
  const container = document.createElement('div')
  container.innerHTML = editorData.value
  return truncateForAIPolish(container.textContent || '', 4000)
}

const readEditorSelectionText = (selection: any) => {
  const parts: string[] = []
  for (const range of selection.getRanges()) {
    for (const item of range.getItems()) {
      if (item.is?.('$textProxy') && item.data) {
        parts.push(item.data)
      }
    }
  }

  return normalizeSelectedText(parts.join(''))
}

const readEditorSelectionHtml = (editor: any, selection: any) => {
  try {
    const fragment = editor.model.getSelectedContent(selection)
    return editor.data.stringify(fragment).trim()
  } catch {
    return ''
  }
}

const rememberEditorSelection = () => {
  const editor = editorInstance.value as any
  if (!editor) return

  const selection = editor.model.document.selection
  const selectedText = readEditorSelectionText(selection)
  const selectedHtml = readEditorSelectionHtml(editor, selection)
  const firstRange = selection.getFirstRange()
  if (!selectedText || !firstRange || firstRange.isCollapsed) {
    if (editor.editing.view.document.isFocused) {
      lastEditorSelectionText.value = ''
      lastEditorSelectionHtml.value = ''
      lastEditorSelectionRange.value = null
    }
    return
  }

  lastEditorSelectionText.value = selectedText
  lastEditorSelectionHtml.value = selectedHtml
  lastEditorSelectionRange.value = firstRange.clone()
}

const getCurrentEditorSelectionSnapshot = () => {
  const editor = editorInstance.value as any
  const selection = editor?.model.document.selection
  const firstRange = selection?.getFirstRange()
  const selectedText = selection ? readEditorSelectionText(selection) : ''
  const selectedHtml = selection ? readEditorSelectionHtml(editor, selection) : ''

  if (selectedText && firstRange && !firstRange.isCollapsed) {
    return {
      range: firstRange.clone(),
      text: selectedText,
      richText: selectedHtml,
      inputFormat: selectedHtml ? 'html' : 'plain_text'
    }
  }

  const fallbackRange = lastEditorSelectionRange.value
  const fallbackText = lastEditorSelectionText.value
  const fallbackHtml = lastEditorSelectionHtml.value
  if (fallbackText && fallbackRange && !fallbackRange.isCollapsed) {
    return {
      range: fallbackRange.clone(),
      text: fallbackText,
      richText: fallbackHtml,
      inputFormat: fallbackHtml ? 'html' : 'plain_text'
    }
  }

  return null
}

const removeAIPolishMarker = () => {
  const editor = editorInstance.value as any
  const markerName = aiPolishMarkerName.value
  if (!editor || !markerName || !editor.model.markers.has(markerName)) {
    aiPolishMarkerName.value = ''
    return
  }

  editor.model.change((writer: any) => {
    writer.removeMarker(markerName)
  })
  aiPolishMarkerName.value = ''
}

const updateAIPolishMarker = (range: any) => {
  const editor = editorInstance.value as any
  if (!editor || !range || range.isCollapsed) return false

  const markerName =
    aiPolishMarkerName.value || `${AI_POLISH_MARKER_GROUP}:${Date.now()}:${aiPolishMarkerSeed++}`

  try {
    editor.model.change((writer: any) => {
      if (editor.model.markers.has(markerName)) {
        writer.updateMarker(markerName, { range, usingOperation: false, affectsData: false })
      } else {
        writer.addMarker(markerName, { range, usingOperation: false, affectsData: false })
      }
    })
    aiPolishMarkerName.value = markerName
    return true
  } catch {
    if (!editor.model.markers.has(markerName)) {
      aiPolishMarkerName.value = ''
    }
    return false
  }
}

const updatePolishTargetFromSelection = (showSuccessToast = true) => {
  const snapshot = getCurrentEditorSelectionSnapshot()
  if (!snapshot) {
    toast.add({
      severity: 'warning',
      summary: '请先选择正文',
      detail: '选中一段文字后再更新替换目标',
      life: 2600
    })
    return null
  }

  if (!updateAIPolishMarker(snapshot.range)) {
    toast.add({
      severity: 'error',
      summary: '替换目标更新失败',
      detail: '请重新选择正文后再试',
      life: 2600
    })
    return null
  }

  lastEditorSelectionRange.value = snapshot.range.clone()
  lastEditorSelectionText.value = snapshot.text
  lastEditorSelectionHtml.value = snapshot.richText

  if (showSuccessToast) {
    toast.add({
      severity: 'success',
      summary: '已更新替换目标',
      detail: '高亮内容将作为 AI 候选的替换范围',
      life: 2200
    })
  }

  return snapshot
}

const closePolishPanel = () => {
  removeAIPolishMarker()
  aiDrawerOpen.value = false
  aiPolishSession.value = null
  polishSuggestions.value = []
  selectedSuggestionIndex.value = -1
}

const requestAIPolish = async (
  mode: AdminAIPolishMode,
  target: AdminAIPolishTarget,
  text: string,
  richContext: { richText?: string; inputFormat?: 'plain_text' | 'html' } = {}
) => {
  const normalizedText = normalizeSelectedText(text)
  const hasContext = Boolean(
    normalizedText ||
    articleTitle.value.trim() ||
    articleSummary.value.trim() ||
    getArticleTextExcerpt().trim()
  )

  if (!hasContext) {
    toast.add({
      severity: 'warning',
      summary: '没有可润色的内容',
      detail: '请先选择正文，或填写标题、摘要、正文内容',
      life: 2600
    })
    return
  }

  isPolishing.value = true
  aiDrawerOpen.value = true
  polishTarget.value = target
  polishMode.value = mode
  if (target !== 'content_selection') {
    removeAIPolishMarker()
  }
  polishSuggestions.value = []
  selectedSuggestionIndex.value = -1
  aiPolishSession.value = createAIPolishSession({
    mode,
    target,
    sourceText: normalizedText,
    sourceRichText: richContext.richText
  })

  try {
    const response = await polishAdminText(
      buildAIPolishRequest({
        mode,
        target,
        text: normalizedText,
        richText: richContext.richText,
        inputFormat: richContext.inputFormat,
        articleId: article.value.id,
        articleTitle: articleTitle.value,
        articleSummary: articleSummary.value,
        articleExcerpt: getArticleTextExcerpt()
      })
    )

    polishSuggestions.value = response.data.suggestions ?? []
    if (aiPolishSession.value) {
      aiPolishSession.value.status = 'ready'
    }
    if (!polishSuggestions.value.length) {
      toast.add({
        severity: 'warning',
        summary: '没有返回候选',
        detail: '可以稍后重试或换一种润色方式',
        life: 2600
      })
    }
  } catch (error) {
    if (aiPolishSession.value) {
      aiPolishSession.value.status = 'error'
    }
    polishSuggestions.value = []
    toast.add({
      severity: 'error',
      summary: 'AI 润色失败',
      detail: getAdminUploadErrorMessage(error, '请稍后重试'),
      life: 3200
    })
  } finally {
    isPolishing.value = false
  }
}

const requestSelectedContentPolish = (mode: AdminAIPolishMode) => {
  const selectionSnapshot = getCurrentEditorSelectionSnapshot()
  if (!selectionSnapshot) {
    toast.add({
      severity: 'warning',
      summary: '请先选择正文',
      detail: '选中一段文字后再使用 AI 润色',
      life: 2600
    })
    return
  }

  if (!updateAIPolishMarker(selectionSnapshot.range)) {
    toast.add({
      severity: 'error',
      summary: '替换目标锁定失败',
      detail: '请重新选择正文后再试',
      life: 2600
    })
    return
  }

  lastEditorSelectionRange.value = selectionSnapshot.range.clone()
  lastEditorSelectionText.value = selectionSnapshot.text
  lastEditorSelectionHtml.value = selectionSnapshot.richText
  void requestAIPolish(mode, 'content_selection', selectionSnapshot.text, {
    richText: selectionSnapshot.richText,
    inputFormat: selectionSnapshot.inputFormat
  })
}

const requestTitleCandidates = () => {
  void requestAIPolish('title_candidates', 'title', articleTitle.value)
}

const requestSummaryCandidates = () => {
  void requestAIPolish('summary_candidates', 'summary', articleSummary.value)
}

const replaceSelectionWithSuggestion = (suggestion: AdminAIPolishSuggestion) => {
  const editor = editorInstance.value as any
  if (!editor) return

  const markerName = aiPolishMarkerName.value
  const marker = markerName ? editor.model.markers.get(markerName) : null
  const range = marker?.getRange()
  if (!range || range.isCollapsed) {
    toast.add({
      severity: 'warning',
      summary: '选区已失效',
      detail: '请重新选择正文，并点击“更新替换目标”后再替换',
      life: 2600
    })
    return
  }

  try {
    editor.model.change((writer: any) => {
      const replacementRange = range.clone()
      writer.removeMarker(markerName)
      insertSuggestionContent(editor, suggestion.content, replacementRange)
    })
    editorData.value = editor.getData()
    aiPolishMarkerName.value = ''
    lastEditorSelectionRange.value = null
    lastEditorSelectionText.value = ''
    lastEditorSelectionHtml.value = ''
    toast.add({ severity: 'success', summary: '已替换选区', detail: '记得保存文章', life: 2200 })
  } catch {
    aiPolishMarkerName.value = ''
    toast.add({
      severity: 'error',
      summary: '替换失败',
      detail: '请重新选择正文，并点击“更新替换目标”后再试',
      life: 2600
    })
  }
}

const applyFieldSuggestion = (suggestion: AdminAIPolishSuggestion) => {
  const content = plainSuggestionText(suggestion.content)
  if (polishTarget.value === 'title') {
    articleTitle.value = content
  } else if (polishTarget.value === 'summary') {
    articleSummary.value = content
  }
  toast.add({ severity: 'success', summary: '已应用候选', detail: '记得保存文章', life: 2200 })
}

const applyPolishSuggestion = (suggestion: AdminAIPolishSuggestion) => {
  if (polishTarget.value === 'content_selection') {
    replaceSelectionWithSuggestion(suggestion)
  } else {
    applyFieldSuggestion(suggestion)
  }
}

const previewPolishSuggestion = (index: number) => {
  selectedSuggestionIndex.value = index
  if (aiPolishSession.value) {
    aiPolishSession.value.selectedSuggestionIndex = index
  }

  const suggestion = polishSuggestions.value[index]
  if (!suggestion) return

  if (polishTarget.value === 'title') {
    openPreview({ title: plainSuggestionText(suggestion.content) })
  } else if (polishTarget.value === 'summary') {
    openPreview({ summary: plainSuggestionText(suggestion.content) })
  } else {
    openPreview({ content: buildContentSelectionPreview(suggestion.content) })
  }
}

const copyPolishSuggestion = async (suggestion: AdminAIPolishSuggestion) => {
  try {
    if (polishTarget.value === 'content_selection') {
      const html = getPolishSuggestionPreviewHtml(suggestion)
      await writeRichClipboard(html, htmlToPlainText(html))
    } else {
      await writeRichClipboard('', plainSuggestionText(suggestion.content))
    }
    toast.add({
      severity: 'success',
      summary: '已复制',
      detail: '候选内容已复制到剪贴板，支持富文本粘贴',
      life: 2200
    })
  } catch {
    toast.add({ severity: 'error', summary: '复制失败', detail: '请手动复制候选内容', life: 2600 })
  }
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
    savedIsPublished.value = isPublished.value
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

const plainSuggestionText = (value: string) => {
  const container = document.createElement('div')
  container.innerHTML = value
  return normalizeSelectedText(container.textContent || value)
}

const buildContentSelectionPreview = (content: string) => {
  const editor = editorInstance.value as any
  const session = aiPolishSession.value
  const sourceText = session?.sourceText || ''
  if (!sourceText) return editorData.value

  const replacementHtml = getSuggestionPreviewHtml(content, editor)
  const richPreview = buildAIPolishContentPreview({
    articleContent: editorData.value,
    sourceText,
    sourceRichText: session?.sourceRichText,
    replacementHtml
  })
  if (richPreview !== editorData.value) return richPreview

  const container = document.createElement('div')
  container.innerHTML = editorData.value
  const replacementTemplate = document.createElement('template')
  replacementTemplate.innerHTML = replacementHtml
  const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT)
  let node = walker.nextNode()
  while (node) {
    const text = node.textContent || ''
    if (text.includes(sourceText)) {
      const fragment = replacementTemplate.content.cloneNode(true)
      if (text === sourceText) {
        node.parentNode?.replaceChild(fragment, node)
      } else {
        const [before, after] = text.split(sourceText)
        const range = document.createRange()
        range.selectNode(node)
        const wrapper = document.createElement('span')
        wrapper.append(before, fragment, after)
        range.deleteContents()
        range.insertNode(wrapper)
      }
      return container.innerHTML
    }
    node = walker.nextNode()
  }
  return editorData.value
}

const getPolishSuggestionPreviewHtml = (suggestion: AdminAIPolishSuggestion) => {
  const editor = editorInstance.value as any
  return getSuggestionPreviewHtml(suggestion.content, editor)
}

const openPreview = (
  overrides: { title?: string; summary?: string; content?: string } | null = null
) => {
  previewOverrides.value = overrides
  previewOpen.value = true
}

const closePreview = () => {
  previewOpen.value = false
  previewOverrides.value = null
}

const updateCategory = (event: Event) => {
  const value = (event.target as HTMLSelectElement).value
  article.value.category_id = value || undefined
}

const goBack = () => {
  void router.push(adminArticleListLocation.value)
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
  removeAIPolishMarker()
  window.removeEventListener('keydown', handleSaveShortcut)
  window.removeEventListener('beforeunload', handleBeforeUnload)
})
</script>

<template>
  <main class="admin-editor-shell">
    <header
      class="sticky top-0 z-30 -mx-4 border-b border-border/70 bg-background/95 px-4 py-2 backdrop-blur sm:-mx-6 sm:px-6 lg:-mx-8 lg:px-8"
    >
      <div
        class="mx-auto flex max-w-7xl flex-col gap-2 xl:flex-row xl:items-center xl:justify-between"
      >
        <div class="flex min-w-0 flex-1 flex-col gap-2 sm:flex-row sm:items-center">
          <AppButton variant="ghost" size="sm" class="w-max" @click="goBack">
            <ArrowLeft class="size-4" aria-hidden="true" />
            返回
          </AppButton>
          <div class="min-w-0 flex-1">
            <h1 class="m-0 truncate text-lg font-black text-foreground">
              {{ articleTitle || '无标题文章' }}
            </h1>
            <p class="m-0 text-xs font-semibold text-muted-foreground">实际标题在文章设置中编辑</p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <AppBadge :tone="savedPublishTone">
            {{ savedPublishText }}
          </AppBadge>
          <AppBadge :tone="saveStatusTone">{{ saveStatusText }}</AppBadge>
          <AppBadge tone="neutral" class="tabular-nums">
            {{ charCount }} 字符 / {{ wordCount }} 词
          </AppBadge>
          <AppButton
            variant="secondary"
            size="sm"
            :disabled="!canUseAIPolish"
            @click="requestSelectedContentPolish('improve')"
          >
            <Sparkles class="size-4" aria-hidden="true" />
            润色
          </AppButton>
          <AppButton
            variant="ghost"
            size="sm"
            :disabled="!canUseAIPolish"
            @click="requestSelectedContentPolish('shorten')"
          >
            精简
          </AppButton>
          <AppButton
            variant="ghost"
            size="sm"
            :disabled="!canUseAIPolish"
            @click="requestSelectedContentPolish('expand')"
          >
            扩写
          </AppButton>
          <AppButton
            variant="secondary"
            size="sm"
            :disabled="!isLayoutReady"
            @click="openPreview()"
          >
            <Eye class="size-4" aria-hidden="true" />
            预览
          </AppButton>
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

    <section v-else class="admin-editor-frame" :class="{ 'has-ai-drawer': aiDrawerOpen }">
      <div class="admin-editor-panel min-w-0 space-y-3">
        <div class="archive-surface admin-editor-content overflow-visible rounded-archive">
          <Ckeditor
            v-model="editorData"
            :editor="ClassicEditor"
            :config="editorConfig"
            @ready="onEditorReady"
          />
        </div>
      </div>

      <div class="admin-editor-settings min-w-0 space-y-3">
        <aside
          v-if="aiDrawerOpen"
          class="admin-ai-drawer archive-surface rounded-archive p-4"
          aria-label="AI 候选"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <h2 class="m-0 text-base font-black text-foreground">{{ polishPanelTitle }}</h2>
              <p class="m-0 mt-1 text-pretty text-xs font-semibold text-muted-foreground">
                {{ isPolishing ? '正在生成候选' : '选择候选，确认后才会改动文章' }}
              </p>
            </div>
            <AppButton
              variant="ghost"
              size="icon"
              aria-label="关闭 AI 候选"
              @click="closePolishPanel"
            >
              <X class="size-4" aria-hidden="true" />
            </AppButton>
          </div>

          <div v-if="isPolishing" class="mt-4 text-sm font-semibold text-muted-foreground">
            正在准备候选...
          </div>

          <div
            v-else-if="aiPolishSession?.status === 'error'"
            class="mt-4 rounded-archive border border-danger/30 bg-danger/10 p-3 text-sm font-semibold text-danger"
          >
            生成失败，请稍后重试
          </div>

          <div v-else class="mt-4 space-y-3">
            <div
              v-if="polishTarget === 'content_selection'"
              class="rounded-archive border border-accent/30 bg-accent/5 p-3"
            >
              <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p class="m-0 text-xs font-semibold leading-5 text-muted-foreground">
                  高亮内容是当前替换目标。重新选择正文后，可以先更新目标再应用候选。
                </p>
                <AppButton
                  variant="secondary"
                  size="sm"
                  class="w-max shrink-0"
                  @mousedown.prevent
                  @click="updatePolishTargetFromSelection()"
                >
                  更新替换目标
                </AppButton>
              </div>
            </div>

            <article
              v-for="(suggestion, index) in polishSuggestions"
              :key="`${suggestion.content}-${index}`"
              class="rounded-archive border border-border bg-surface p-3"
              :class="selectedSuggestionIndex === index ? 'border-accent bg-accent/5' : ''"
            >
              <div
                v-if="polishTarget === 'content_selection'"
                class="reading-prose reading-prose--compact admin-ai-suggestion-content ck-content"
                v-html="getPolishSuggestionPreviewHtml(suggestion)"
              />
              <p v-else class="m-0 whitespace-pre-wrap text-pretty text-sm leading-7 text-foreground">
                {{ plainSuggestionText(suggestion.content) }}
              </p>
              <p
                v-if="suggestion.reason"
                class="m-0 mt-2 text-xs font-semibold text-muted-foreground"
              >
                {{ suggestion.reason }}
              </p>
              <div class="mt-3 flex flex-wrap gap-2">
                <AppButton size="sm" @click="applyPolishSuggestion(suggestion)">
                  <Sparkles class="size-4" aria-hidden="true" />
                  {{ getAIPolishApplyLabel(polishTarget) }}
                </AppButton>
                <AppButton variant="secondary" size="sm" @click="previewPolishSuggestion(index)">
                  预览
                </AppButton>
                <AppButton variant="secondary" size="sm" @click="copyPolishSuggestion(suggestion)">
                  <Copy class="size-4" aria-hidden="true" />
                  复制
                </AppButton>
              </div>
            </article>
          </div>
        </aside>

        <aside class="archive-surface h-max rounded-archive p-4">
          <h2 class="m-0 text-base font-black text-foreground">文章设置</h2>
          <div class="mt-4 space-y-4">
            <div class="block space-y-2">
              <div class="flex items-center justify-between gap-3">
                <label for="admin-article-title" class="text-sm font-bold text-foreground"
                  >标题</label
                >
                <AppButton
                  variant="ghost"
                  size="sm"
                  class="w-max"
                  :disabled="!canUseAIPolish"
                  @click.stop="requestTitleCandidates"
                >
                  <Wand2 class="size-4" aria-hidden="true" />
                  标题候选
                </AppButton>
              </div>
              <AppInput
                id="admin-article-title"
                v-model="articleTitle"
                aria-label="文章标题"
                placeholder="无标题文章"
                class="rounded-archive"
                :disabled="!isLayoutReady"
              />
            </div>

            <div class="block space-y-2">
              <div class="flex items-center justify-between gap-3">
                <label for="admin-article-summary" class="text-sm font-bold text-foreground"
                  >摘要</label
                >
                <AppButton
                  variant="ghost"
                  size="sm"
                  class="w-max"
                  :disabled="!canUseAIPolish"
                  @click.stop="requestSummaryCandidates"
                >
                  <Wand2 class="size-4" aria-hidden="true" />
                  摘要候选
                </AppButton>
              </div>
              <textarea
                id="admin-article-summary"
                v-model="articleSummary"
                rows="5"
                class="w-full resize-y rounded-archive border border-border bg-surface px-4 py-3 text-sm leading-6 text-foreground outline-none transition-colors placeholder:text-muted-foreground/80 focus:border-accent focus:ring-2 focus:ring-accent/20"
                placeholder="简短说明这篇文章解决什么问题"
              />
            </div>

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
      </div>
    </section>

    <Teleport to="body">
      <div
        v-if="previewOpen"
        class="fixed inset-0 z-50 overflow-y-auto bg-background/72 p-4 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        aria-labelledby="admin-article-preview-title"
        @click.self="closePreview"
      >
        <div class="mx-auto my-6 w-full max-w-[860px]">
          <div class="mb-3 flex items-center justify-between gap-3">
            <div>
              <p class="m-0 text-xs font-bold text-muted-foreground">实际阅读效果</p>
              <h2 id="admin-article-preview-title" class="m-0 text-xl font-black text-foreground">
                文章预览
              </h2>
            </div>
            <AppButton variant="secondary" size="icon" aria-label="关闭预览" @click="closePreview">
              <X class="size-4" aria-hidden="true" />
            </AppButton>
          </div>

          <article class="archive-surface rounded-[1.1rem] p-5 md:p-8">
            <header class="space-y-5 border-b border-border pb-6">
              <div class="flex flex-wrap gap-2">
                <AppBadge v-if="article.category_id" tone="accent">
                  {{
                    categories.find(
                      (category) => String(category.id) === String(article.category_id)
                    )?.name || '分类'
                  }}
                </AppBadge>
                <AppBadge :tone="isPublished ? 'accent' : 'neutral'">
                  {{ isPublished ? '已发布' : '草稿' }}
                </AppBadge>
              </div>
              <h1 class="m-0 text-3xl font-extrabold leading-tight text-foreground md:text-4xl">
                {{ previewTitle || '无标题文章' }}
              </h1>
            </header>

            <section
              v-if="previewSummary"
              class="my-6 rounded-archive border border-border bg-surface-raised p-4"
            >
              <p class="m-0 text-xs font-black text-muted-foreground">摘要</p>
              <p class="m-0 mt-2 text-base leading-8 text-foreground/85">{{ previewSummary }}</p>
            </section>

            <img
              v-if="coverPreview"
              :src="coverPreview"
              :alt="previewTitle || '文章封面'"
              class="mb-6 aspect-[16/9] w-full rounded-archive object-cover"
            />
            <div class="reading-prose ck-content admin-preview-content" v-html="previewContent" />
          </article>
        </div>
      </div>
    </Teleport>
  </main>
</template>
