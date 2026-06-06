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
        <el-button type="primary" :loading="isSaving" @click="saveArticle">保存文章</el-button>
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
                :on-change="handleCoverChange"
              >
                <el-icon class="uploader-icon"><Plus /></el-icon>
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
      <div class="last-save" v-if="lastSaveTime">上次保存: {{ lastSaveTime }}</div>
    </div>
  </div>
</template>

<script setup lang="ts" name="articleEditor">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ClassicEditor } from 'ckeditor5'
import { Ckeditor } from '@ckeditor/ckeditor5-vue'
import 'ckeditor5/ckeditor5.css'

import { Plus, Delete, FullScreen } from '@element-plus/icons-vue'
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
const lastSaveTime = ref('')

const STORAGE_KEY = 'nostalgia_article_draft'

let initialContent = ''
let initialArticle = ''

// 从本地存储恢复
const restoreDraft = () => {
  const draft = sessionStorage.getItem(STORAGE_KEY)
  if (draft) {
    const data = JSON.parse(draft)
    if (data.id === article.value.id) {
      // 只有当草稿内容和当前接口拉取的内容真正不一致时，才进行恢复
      if (
        data.content !== editorData.value ||
        JSON.stringify(data.article) !== JSON.stringify(article.value)
      ) {
        editorData.value = data.content
        Object.assign(article.value, data.article)
        ElMessage.info('已恢复未保存的内容')
      } else {
        // 如果内容完全一致，说明不需要恢复，顺手清掉无用草稿
        clearDraft()
      }
    }
  }
}

// 自动保存草稿到 session
watch(
  [editorData, article],
  () => {
    if (editorData.value === initialContent && JSON.stringify(article.value) === initialArticle)
      return

    sessionStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({
        id: article.value.id,
        content: editorData.value,
        article: article.value,
      }),
    )
  },
  { deep: true },
)

// 保存成功后清除草稿
const clearDraft = () => {
  sessionStorage.removeItem(STORAGE_KEY)
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
      initialArticle = JSON.stringify(article.value)
    } else {
      isNew = true
      // Create a draft first
      const res = await createArticleApi({ title: '无标题文章', is_publish: false })
      const oldPath = route.fullPath
      article.value = { ...res.data.article }
      initialContent = editorData.value // It's empty initially
      initialArticle = JSON.stringify(article.value)
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
  } catch (error) {
    ElMessage.error('加载失败')
  }
}

const saveArticle = async () => {
  if (!article.value.title) return ElMessage.warning('标题不能为空')
  isSaving.value = true
  try {
    const params = {
      ...article.value,
      content: editorData.value,
    }
    await updateArticleApi(params)
    lastSaveTime.value = dayjs().format('HH:mm:ss')
    ElMessage.success('保存成功')
    clearDraft()
  } finally {
    isSaving.value = false
  }
}

const handleCoverChange = async (file: any) => {
  const reader = new FileReader()
  reader.readAsDataURL(file.raw)
  reader.onload = async () => {
    const result = reader.result as string
    const base64 = result.split(',')[1]
    try {
      const res = await http.post<any>('/util/upload_file', {
        article_id: article.value.id,
        content: base64,
        type: 'cover',
      })
      article.value.cover = res.data.url + '?t=' + Date.now()
      ElMessage.success('封面已更新')
    } catch (e) {
      ElMessage.error('上传失败')
    }
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

onMounted(() => {
  initData()
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
  background-color: #f5f7f9;
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
    background-color: #ffffff;
    border-bottom: 1px solid #dcdfe6;
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
          max-width: 850px;
          margin-bottom: 50px;
          background-color: #ffffff;
          box-shadow: 0 2px 12px 0 rgb(0 0 0 / 10%);
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
            background-color: #fafafa !important;
            border: none !important;
            border-bottom: 1px solid #eeeeee !important;
            box-shadow: 0 2px 4px rgb(0 0 0 / 5%);
          }
        }
      }
    }
    .editor-side {
      width: 320px;
      padding: 20px;
      overflow-y: auto;
      background-color: #ffffff;
      border-left: 1px solid #dcdfe6;
      .w-full {
        width: 100%;
      }
      .cover-preview {
        position: relative;
        width: 100%;
        height: 160px;
        overflow: hidden;
        border: 1px solid #dcdfe6;
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
        border: 1px dashed #dcdfe6;
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
    color: #909399;
    background-color: #ffffff;
    border-top: 1px solid #dcdfe6;
  }
}
</style>
