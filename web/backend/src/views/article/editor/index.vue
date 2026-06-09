<template>
  <div class="editor-container card" :class="{ 'full-screen': isFullScreen }">
    <div class="editor-header">
      <el-input v-model="article.title" placeholder="请输入文章标题" class="title-input" />
      <div class="action-btns">
        <el-button @click="toggleFullScreen">
          <el-icon><FullScreen v-if="!isFullScreen" /><FullScreen v-else /></el-icon>
          {{ isFullScreen ? '退出全屏' : '全屏编辑' }}
        </el-button>
        <el-button @click="goBack">返回</el-button>
        <el-button type="primary" :loading="isSaving" :disabled="isSaving" @click="saveArticle">
          保存文章
        </el-button>
      </div>
    </div>

    <div class="editor-main">
      <!-- 左侧编辑器 -->
      <div class="editor-wrapper paper-style" ref="scrollContainer">
        <ckeditor
          v-if="isLayoutReady"
          v-model="editorData"
          :editor="ClassicEditor"
          :config="finalConfig"
          @ready="onEditorReady"
        />
      </div>

      <!-- 右侧设置栏 -->
      <div class="editor-side" v-show="!isFullScreen">
        <el-card shadow="never" header="文章设置">
          <el-form :model="article" label-position="top">
            <el-form-item label="摘要">
              <el-input
                type="textarea"
                v-model="article.summary"
                :rows="4"
                placeholder="简短的介绍..."
              />
            </el-form-item>
            <el-form-item label="短标识 (Slug)">
              <el-input v-model="article.slug" placeholder="url-slug" />
            </el-form-item>
            <el-form-item label="分类">
              <el-select v-model="article.category_id" placeholder="选择分类" class="w-full">
                <el-option
                  v-for="item in categoryList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="发布状态">
              <el-switch v-model="article.is_publish" active-text="已发布" inactive-text="草稿" />
            </el-form-item>
            <el-form-item label="封面图">
              <div v-if="article.cover" class="cover-preview">
                <el-image :src="article.cover" fit="cover" />
                <el-button
                  type="danger"
                  :icon="Delete"
                  circle
                  class="delete-btn"
                  @click="removeCover"
                />
              </div>
              <el-upload
                v-else
                class="cover-uploader"
                action=""
                :auto-upload="false"
                :show-file-list="false"
                :disabled="isCoverUploading"
                :on-change="handleCoverChange"
              >
                <el-icon class="uploader-icon"
                  ><Loading v-if="isCoverUploading" /><Plus v-else
                /></el-icon>
              </el-upload>
            </el-form-item>
          </el-form>
        </el-card>
      </div>
    </div>

    <div class="editor-footer" v-if="isLayoutReady">
      <div class="word-count">
        <span>字符数: {{ charCount }}</span>
        <span style="margin-left: 15px">单词数: {{ wordCount }}</span>
      </div>
      <div class="save-meta">
        <span class="save-status" :class="saveStatusClass">{{ saveStatusText }}</span>
        <span v-if="hasDraft">已缓存草稿</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts" name="articleEditor">
import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ClassicEditor } from 'ckeditor5'
import { Ckeditor } from '@ckeditor/ckeditor5-vue'
import 'ckeditor5/ckeditor5.css'
import '@/styles/content.scss'

import { Plus, Delete, FullScreen, Loading } from '@element-plus/icons-vue'
import { editorConfig } from '@/config/editorConfig'
import MyUploadAdapter from '@/utils/uploadAdapter'
import { getArticleByIdApi, createArticleApi, updateArticleApi } from '@/api/modules/articles'
import { getCategoryListApi } from '@/api/modules/categories'
import { ElMessage } from 'element-plus'
import http from '@/api'
import { Nostalgia } from '@/api/interface/nostalgia'
import { useTabsStore } from '@/stores/modules/tabs'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()

const isFullScreen = ref(false)
const isLayoutReady = ref(false)
const editorData = ref('')
const article = ref<Partial<Nostalgia.Article>>({
  id: '',
  title: '',
  summary: '',
  content: '',
  category_id: undefined,
  is_publish: false,
  cover: '',
  slug: '',
  check_outdated: false,
})

const tabsStore = useTabsStore()
const categoryList = ref<Nostalgia.Category[]>([])
const isSaving = ref(false)
const isCoverUploading = ref(false)
const lastSaveTime = ref('')
const saveStatus = ref<'saved' | 'dirty' | 'saving' | 'error'>('saved')
const hasDraft = ref(false)
const isTrackingChanges = ref(false)

const STORAGE_KEY = 'nostalgia_article_draft'

let initialContent = ''
let initialArticle = ''

const serializeArticle = () => JSON.stringify(article.value)

const hasUnsavedChanges = () =>
  isTrackingChanges.value &&
  (editorData.value !== initialContent || serializeArticle() !== initialArticle)

const saveStatusText = computed(() => {
  if (saveStatus.value === 'saving') return '保存中...'
  if (saveStatus.value === 'error') return '保存失败，修改仍已缓存'
  if (hasUnsavedChanges()) return hasDraft.value ? '有未保存草稿' : '有未保存修改'
  if (lastSaveTime.value) return `已保存 ${lastSaveTime.value}`
  return '已保存'
})

const saveStatusClass = computed(() => ({
  'is-dirty': hasUnsavedChanges() && saveStatus.value !== 'saving',
  'is-saving': saveStatus.value === 'saving',
  'is-error': saveStatus.value === 'error',
}))

const validateImageFile = (file?: File) => {
  if (!file) return '请选择要上传的图片'
  if (!file.type.startsWith('image/')) return '只能上传图片文件'
  return ''
}

const fileToBase64 = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => {
      const result = reader.result as string
      resolve(result.split(',')[1] || '')
    }
    reader.onerror = () => reject(new Error('读取图片失败'))
  })
}

// 从本地存储恢复
const restoreDraft = () => {
  const draft = sessionStorage.getItem(STORAGE_KEY)
  if (!draft) return

  try {
    const data = JSON.parse(draft)
    if (data.id === article.value.id) {
      // 只有当草稿内容和当前接口拉取的内容真正不一致时，才进行恢复
      if (
        data.content !== editorData.value ||
        JSON.stringify(data.article) !== JSON.stringify(article.value)
      ) {
        editorData.value = data.content
        Object.assign(article.value, data.article)
        hasDraft.value = true
        saveStatus.value = 'dirty'
        ElMessage.info('已恢复未保存的内容')
      } else {
        // 如果内容完全一致，说明不需要恢复，顺手清掉无用草稿
        clearDraft()
      }
    }
  } catch (error) {
    clearDraft()
  }
}

// 自动保存草稿到 session
watch(
  [editorData, article],
  () => {
    if (!isTrackingChanges.value) return

    if (!hasUnsavedChanges()) {
      saveStatus.value = 'saved'
      clearDraft()
      return
    }

    sessionStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({
        id: article.value.id,
        content: editorData.value,
        article: article.value,
      }),
    )
    hasDraft.value = true
    if (saveStatus.value !== 'saving' && saveStatus.value !== 'error') {
      saveStatus.value = 'dirty'
    }
  },
  { deep: true },
)

// 保存成功后清除草稿
const clearDraft = () => {
  sessionStorage.removeItem(STORAGE_KEY)
  hasDraft.value = false
}

const charCount = ref(0)
const wordCount = ref(0)

const finalConfig = computed(() => ({
  ...editorConfig,
  extraPlugins: [CustomUploadAdapterPlugin, WordCountPlugin],
}))

function WordCountPlugin(editor: any) {
  const wordCountPlugin = editor.plugins.get('WordCount')
  wordCountPlugin.on('update', (evt: any, stats: any) => {
    charCount.value = stats.characters
    wordCount.value = stats.words
  })
}

function CustomUploadAdapterPlugin(editor: any) {
  editor.plugins.get('FileRepository').createUploadAdapter = (loader: any) => {
    return new MyUploadAdapter(loader, article.value.id as string, 'content')
  }
}

const onEditorReady = (editorInstance: any) => {
  editorInstance.ui.view.editable.element?.classList.add('nostalgia-content')
}

const initData = async () => {
  try {
    const catRes = await getCategoryListApi()
    categoryList.value = catRes.data.list

    const id = route.query.id as string
    let isNew = false
    if (id) {
      const res = await getArticleByIdApi(id)
      article.value = { ...res.data.article }
      editorData.value = res.data.article.content || ''
      initialContent = editorData.value
      initialArticle = serializeArticle()
    } else {
      isNew = true
      // Create a draft first
      const res = await createArticleApi({ title: '无标题文章', is_publish: false })
      const oldPath = route.fullPath
      article.value = { ...res.data.article }
      initialContent = editorData.value // It's empty initially
      initialArticle = serializeArticle()
      const newPath = `/article/editor?id=${article.value.id}`

      // 更新 TabsStore
      const tabIndex = tabsStore.tabsMenuList.findIndex((item) => item.path === oldPath)
      if (tabIndex !== -1) {
        tabsStore.tabsMenuList[tabIndex].path = newPath
      }

      router.replace(newPath)
    }
    isLayoutReady.value = true
    if (!isNew) {
      restoreDraft()
    }
    isTrackingChanges.value = true
    saveStatus.value = hasUnsavedChanges() ? 'dirty' : 'saved'
  } catch (error) {
    ElMessage.error('加载失败')
  }
}

const saveArticle = async () => {
  if (isSaving.value) return
  if (!article.value.title) return ElMessage.warning('标题不能为空')
  isSaving.value = true
  saveStatus.value = 'saving'
  try {
    const params = {
      ...article.value,
      content: editorData.value,
    }
    await updateArticleApi(params)
    initialContent = editorData.value
    initialArticle = serializeArticle()
    lastSaveTime.value = dayjs().format('HH:mm:ss')
    saveStatus.value = 'saved'
    ElMessage.success('保存成功')
    clearDraft()
  } catch (error: any) {
    saveStatus.value = 'error'
    ElMessage.error(error?.response?.data?.error || '保存失败，请稍后重试')
  } finally {
    isSaving.value = false
  }
}

const handleCoverChange = async (file: any) => {
  const rawFile = file.raw as File | undefined
  const validationMessage = validateImageFile(rawFile)
  if (validationMessage) {
    ElMessage.warning(validationMessage)
    return
  }

  isCoverUploading.value = true
  try {
    const base64 = await fileToBase64(rawFile!)
    const res = await http.post<any>('/util/upload_file', {
      article_id: article.value.id,
      content: base64,
      type: 'cover',
    })
    article.value.cover = res.data.url + '?t=' + Date.now()
    ElMessage.success('封面已更新')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error || error?.message || '封面上传失败')
  } finally {
    isCoverUploading.value = false
  }
}

const removeCover = () => {
  article.value.cover = ''
}

const goBack = () => {
  router.push({ name: 'articleManage' })
}

const toggleFullScreen = () => {
  isFullScreen.value = !isFullScreen.value
}

const handleSaveShortcut = (event: KeyboardEvent) => {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
    event.preventDefault()
    saveArticle()
  }
}

const handleBeforeUnload = (event: BeforeUnloadEvent) => {
  if (!hasUnsavedChanges() || isSaving.value) return
  event.preventDefault()
  event.returnValue = ''
}

onBeforeRouteLeave((_to, _from, next) => {
  if (!hasUnsavedChanges()) {
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
  initData()
  window.addEventListener('keydown', handleSaveShortcut)
  window.addEventListener('beforeunload', handleBeforeUnload)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleSaveShortcut)
  window.removeEventListener('beforeunload', handleBeforeUnload)
})
</script>

<style scoped lang="scss">
.editor-container {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 0;
  overflow: hidden;
  background-color: var(--el-fill-color-extra-light);
  transition: all 0.3s ease;
  &.full-screen {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 2000;
    width: 100vw;
    height: 100vh;
  }
  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 20px;
    background-color: var(--el-bg-color);
    border-bottom: 1px solid var(--el-border-color-light);
    .title-input {
      font-size: 20px;
      font-weight: bold;
      :deep(.el-input__wrapper) {
        box-shadow: none !important;
      }
    }
    .action-btns {
      display: flex;
      gap: 10px;
    }
  }
  .editor-main {
    display: flex;
    flex: 1;
    overflow: hidden;
    .editor-wrapper {
      display: flex;
      flex: 1;
      flex-direction: column;
      align-items: center;
      padding: 30px;
      overflow-y: hidden; /* 移除外层滚动条，让 CKEditor 内部滚动 */
      background-color: var(--el-fill-color-extra-light);
      &.paper-style {
        :deep(.ck-editor) {
          display: flex;
          flex-direction: column;
          width: 100%;
          max-width: 820px;
          margin-bottom: 50px;
          background-color: var(--el-bg-color);
          border: 1px solid var(--el-border-color-lighter);
          border-radius: 12px;
          box-shadow: 0 14px 38px rgb(0 0 0 / 8%);
          .ck-editor__editable {
            height: calc(100vh - 200px); /* 限制高度触发内部滚动 */
            min-height: 400px;
            padding: 40px 60px;
            border: none !important;
          }
          .ck-toolbar {
            position: sticky;
            top: 0;
            z-index: 10;
            background-color: var(--el-bg-color-overlay) !important;
            border: none !important;
            border-bottom: 1px solid var(--el-border-color-light) !important;
            box-shadow: 0 2px 4px rgb(0 0 0 / 5%);
          }
        }
      }
    }
    .editor-side {
      width: 320px;
      padding: 20px;
      overflow-y: auto;
      background-color: var(--el-bg-color);
      border-left: 1px solid var(--el-border-color-light);
      .w-full {
        width: 100%;
      }
      .cover-preview {
        position: relative;
        width: 100%;
        height: 160px;
        overflow: hidden;
        border: 1px solid var(--el-border-color);
        border-radius: 4px;
        .el-image {
          width: 100%;
          height: 100%;
        }
        .delete-btn {
          position: absolute;
          top: 5px;
          right: 5px;
          opacity: 0.8;
          &:hover {
            opacity: 1;
          }
        }
      }
      .cover-uploader {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 100%;
        height: 160px;
        cursor: pointer;
        border: 1px dashed var(--el-border-color);
        border-radius: 4px;
        &:hover {
          border-color: #409eff;
        }
        .uploader-icon {
          font-size: 28px;
          color: #8c939d;
        }
      }
    }
  }
  .editor-footer {
    display: flex;
    justify-content: space-between;
    padding: 8px 20px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    background-color: var(--el-bg-color);
    border-top: 1px solid var(--el-border-color-light);

    .save-meta {
      display: flex;
      gap: 12px;
      align-items: center;
    }

    .save-status {
      color: var(--el-color-success);

      &.is-dirty {
        color: var(--el-color-warning);
      }

      &.is-saving {
        color: var(--el-color-primary);
      }

      &.is-error {
        color: var(--el-color-danger);
      }
    }
  }
}

@media (max-width: 960px) {
  .editor-container {
    .editor-main {
      .editor-wrapper {
        padding: 16px;

        &.paper-style {
          :deep(.ck-editor .ck-editor__editable) {
            padding: 28px 24px;
          }
        }
      }

      .editor-side {
        width: 280px;
      }
    }
  }
}
</style>
