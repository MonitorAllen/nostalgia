<template>
  <div class="article-manage-container">
    <!-- 顶部操作栏 -->
    <div class="operation-header">
      <div class="header-left">
        <el-button type="primary" :icon="CirclePlus" @click="navToEditor()">发布文章</el-button>
      </div>
      <div class="header-right">
        <el-input
          v-model="queryParams.title"
          placeholder="搜索文章标题..."
          clearable
          class="search-input"
          @keyup.enter="resetAndFetch"
          @clear="resetAndFetch"
        >
          <template #append>
            <el-button :icon="Search" @click="resetAndFetch" />
          </template>
        </el-input>
      </div>
    </div>

    <!-- 卡片列表区域 (支持无限滚动) -->
    <div
      class="article-list-wrapper"
      v-infinite-scroll="loadMore"
      :infinite-scroll-immediate="true"
      :infinite-scroll-delay="300"
      :infinite-scroll-disabled="loading || finished"
      :infinite-scroll-distance="20"
    >
      <div class="article-card-container">
        <div v-for="item in articleList" :key="item.id" class="article-card-item">
          <el-card :body-style="{ padding: '0px' }" shadow="hover">
            <div class="card-content-wrapper">
              <div class="card-cover">
                <el-image :src="item.cover || defaultCover" fit="cover">
                  <template #error>
                    <div class="image-slot">
                      <el-icon><Picture /></el-icon>
                    </div>
                  </template>
                </el-image>
                <div class="publish-status">
                  <el-tag :type="item.is_publish ? 'success' : 'info'" size="small">
                    {{ item.is_publish ? '已发布' : '草稿' }}
                  </el-tag>
                </div>
              </div>
              <div class="card-info">
                <div class="card-header-top">
                  <h3 class="title" @click="navToEditor(item.id)">{{ item.title }}</h3>
                  <div class="category">
                    <el-tag size="small" effect="plain">{{ item.category_name }}</el-tag>
                  </div>
                </div>
                <p class="summary">{{ item.summary || '暂无摘要...' }}</p>
                <div class="card-footer">
                  <div class="meta-info">
                    <span>
                      <el-icon><View /></el-icon>
                      {{ item.views }}
                    </span>
                    <span>
                      <el-icon><Pointer /></el-icon>
                      {{ item.likes }}
                    </span>
                    <span>
                      <el-icon><Calendar /></el-icon>
                      {{ formatTime(item.created_at, 'YYYY-MM-DD') }}
                    </span>
                  </div>
                  <div class="card-actions">
                    <el-dropdown trigger="click">
                      <el-button type="primary" link :icon="MoreFilled" />
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item :icon="EditPen" @click="navToEditor(item.id)"
                            >编辑内容</el-dropdown-item
                          >
                          <el-dropdown-item :icon="Edit" @click="openEditInfo(item)"
                            >编辑信息</el-dropdown-item
                          >
                          <el-dropdown-item :icon="Promotion" @click="changePublishStatus(item)">
                            {{ item.is_publish ? '设为草稿' : '发布文章' }}
                          </el-dropdown-item>
                          <el-dropdown-item
                            :icon="Delete"
                            divided
                            @click="deleteArticle(item)"
                            class="delete-text"
                          >
                            删除
                          </el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </div>
      </div>

      <!-- 加载状态提示 -->
      <div class="loading-status">
        <el-divider v-if="finished">没有更多了</el-divider>
        <div v-else-if="loading" class="loading-icon">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载中...</span>
        </div>
      </div>
    </div>

    <!-- 编辑文章信息对话框 -->
    <el-dialog v-model="dialogVisible" title="编辑文章信息" width="500px">
      <el-form :model="currentArticle" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="currentArticle.title" />
        </el-form-item>
        <el-form-item label="摘要">
          <el-input v-model="currentArticle.summary" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="短标识">
          <el-input v-model="currentArticle.slug" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="currentArticle.category_id" placeholder="请选择分类">
            <el-option
              v-for="item in categoryList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="检查时效">
          <el-switch v-model="currentArticle.check_outdated" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveArticleInfo">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts" name="articleManage">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  CirclePlus,
  Delete,
  Edit,
  EditPen,
  View,
  Pointer,
  Calendar,
  Picture,
  MoreFilled,
  Promotion,
  Search,
  Loading,
} from '@element-plus/icons-vue'
import { getArticleListApi, deleteArticleApi, updateArticleApi } from '@/api/modules/articles'
import { getCategoryListApi } from '@/api/modules/categories'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Nostalgia } from '@/api/interface/nostalgia'
import { formatTime } from '@/utils/date'
import defaultCover from '@/assets/images/default-cover.png'

const router = useRouter()

// 数据列表与状态
const articleList = ref<Nostalgia.Article[]>([])
const loading = ref(false)
const finished = ref(false)
const queryParams = reactive({
  title: '',
  page: 1,
  limit: 24,
})

const categoryList = ref<Nostalgia.Category[]>([])

// 获取分类列表
const fetchCategories = async () => {
  const res = await getCategoryListApi()
  categoryList.value = res.data.list
}

// 获取文章列表核心逻辑
const fetchArticleList = async (isAppend = true) => {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getArticleListApi({
      title: queryParams.title,
      page: queryParams.page,
      limit: queryParams.limit,
    })
    const list = res.data.list || []
    const total = res.data.total || 0

    if (isAppend) {
      articleList.value.push(...list)
    } else {
      articleList.value = list
    }

    // 判断是否加载完成
    if (articleList.value.length >= total || list.length === 0) {
      finished.value = true
    } else {
      finished.value = false
    }
  } catch (error) {
    ElMessage.error('获取文章列表失败')
  } finally {
    loading.value = false
  }
}

// 无限滚动加载
const loadMore = () => {
  if (finished.value || loading.value) return
  queryParams.page++
  fetchArticleList(true)
}

// 重置搜索
const resetAndFetch = () => {
  queryParams.page = 1
  articleList.value = []
  finished.value = false
  fetchArticleList(false)
}

onMounted(() => {
  fetchCategories()
  fetchArticleList(false)
})

// 跳转到编辑器
const navToEditor = (id?: string) => {
  if (id) {
    router.push({ name: 'articleEditor', query: { id } })
  } else {
    router.push({ name: 'articleEditor' })
  }
}

// 切换发布状态
const changePublishStatus = async (row: Nostalgia.Article) => {
  try {
    const newStatus = !row.is_publish
    await updateArticleApi({
      id: row.id,
      is_publish: newStatus,
    })
    row.is_publish = newStatus
    ElMessage.success(`文章【${row.title}】发布状态已更新`)
  } catch (error) {
    // 状态保持不变
  }
}

// 编辑文章基本信息
const dialogVisible = ref(false)
const currentArticle = ref<Partial<Nostalgia.Article>>({})
const openEditInfo = (row: Nostalgia.Article) => {
  currentArticle.value = { ...row }
  dialogVisible.value = true
}

const saveArticleInfo = async () => {
  await updateArticleApi(currentArticle.value)
  ElMessage.success('保存成功')
  dialogVisible.value = false
  // 更新本地列表数据
  const index = articleList.value.findIndex((item) => item.id === currentArticle.value.id)
  if (index !== -1) {
    Object.assign(articleList.value[index], currentArticle.value)
  }
}

// 删除文章
const deleteArticle = async (row: Nostalgia.Article) => {
  await ElMessageBox.confirm('确认删除该文章吗？数据不可恢复！', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
  await deleteArticleApi(row.id)
  ElMessage.success('删除成功')
  // 从列表移除
  articleList.value = articleList.value.filter((item) => item.id !== row.id)
}
</script>

<style scoped lang="scss">
.article-manage-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 20px;
  overflow: hidden;
  background-color: var(--el-bg-color-page);
  .operation-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 15px 20px;
    margin-bottom: 20px;
    background-color: var(--el-bg-color);
    border-radius: 8px;
    box-shadow: 0 2px 12px 0 rgb(0 0 0 / 5%);
    .search-input {
      width: 300px;
    }
  }
  .article-list-wrapper {
    flex: 1;
    padding-bottom: 20px;
    overflow-y: auto;
    &::-webkit-scrollbar {
      width: 6px;
    }
    &::-webkit-scrollbar-thumb {
      background: var(--el-border-color-darker);
      border-radius: 3px;
    }
  }
  .article-card-container {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
    gap: 20px;
  }
  .article-card-item {
    .card-content-wrapper {
      display: flex;
      flex-direction: column;
      height: 100%;
    }
    .card-cover {
      position: relative;
      width: 100%;
      height: 180px;
      overflow: hidden;
      background-color: var(--el-fill-color-light);
      .el-image {
        width: 100%;
        height: 100%;
        transition: transform 0.3s ease;
      }
      &:hover .el-image {
        transform: scale(1.05);
      }
      .publish-status {
        position: absolute;
        top: 10px;
        right: 10px;
        z-index: 10;
      }
      .image-slot {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 100%;
        height: 100%;
        font-size: 30px;
        color: var(--el-text-color-secondary);
      }
    }
    .card-info {
      display: flex;
      flex: 1;
      flex-direction: column;
      padding: 16px;
      .card-header-top {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        margin-bottom: 8px;
        .title {
          display: -webkit-box;
          margin: 0;
          overflow: hidden;
          font-size: 18px;
          font-weight: 600;
          color: var(--el-text-color-primary);
          -webkit-line-clamp: 2;
          cursor: pointer;
          -webkit-box-orient: vertical;
          &:hover {
            color: var(--el-color-primary);
          }
        }
      }
      .summary {
        display: -webkit-box;
        margin: 0 0 16px;
        overflow: hidden;
        font-size: 14px;
        line-height: 1.6;
        color: var(--el-text-color-regular);
        -webkit-box-orient: vertical;
        -webkit-line-clamp: 2;
      }
      .card-footer {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-top: auto;
        .meta-info {
          display: flex;
          gap: 15px;
          font-size: 12px;
          color: var(--el-text-color-secondary);
          span {
            display: flex;
            gap: 4px;
            align-items: center;
          }
        }
      }
    }
  }
  .loading-status {
    display: flex;
    justify-content: center;
    padding: 20px 0;
    .loading-icon {
      display: flex;
      gap: 8px;
      align-items: center;
      color: var(--el-text-color-secondary);
    }
  }
}
.delete-text {
  color: var(--el-color-danger) !important;
}

@media screen and (width >= 1200px) {
  .article-card-container {
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  }
}
</style>
