<template>
  <div class="home card">
    <div class="home-header">
      <div class="welcome-text">
        <h2>欢迎回来, {{ userInfo.name }}!</h2>
        <p>这是你的个人博客后台管理系统，今天也要加油哦！</p>
      </div>
      <img src="@/assets/images/welcome.png" alt="welcome" />
    </div>
    <div class="statistics-box">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="statistic-item">
            <div class="item-left">
              <el-icon><Document /></el-icon>
            </div>
            <div class="item-right">
              <div class="label">文章总数</div>
              <div class="value">{{ stats.articleCount }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="statistic-item">
            <div class="item-left">
              <el-icon><View /></el-icon>
            </div>
            <div class="item-right">
              <div class="label">总访问量</div>
              <div class="value">{{ stats.totalViews }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="statistic-item">
            <div class="item-left">
              <el-icon><Collection /></el-icon>
            </div>
            <div class="item-right">
              <div class="label">分类数量</div>
              <div class="value">{{ stats.categoryCount }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="statistic-item">
            <div class="item-left">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="item-right">
              <div class="label">留言评论</div>
              <div class="value">{{ stats.commentCount }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
    <div class="home-main">
      <el-card shadow="never" header="最近发布的文章">
        <el-table :data="recentArticles" style="width: 100%">
          <el-table-column prop="title" label="标题" show-overflow-tooltip />
          <el-table-column prop="category_name" label="分类" width="120" />
          <el-table-column prop="created_at" label="发布时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.is_publish ? 'success' : 'info'">
                {{ scope.row.is_publish ? '已发布' : '草稿' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts" name="home">
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/modules/user'
import { getArticleListApi } from '@/api/modules/articles'
import { getCategoryListApi } from '@/api/modules/categories'
import dayjs from 'dayjs'

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)

const stats = ref({
  articleCount: 0,
  totalViews: 0,
  categoryCount: 0,
  commentCount: 0,
})

const recentArticles = ref([])

const fetchData = async () => {
  try {
    // Fetch articles for stats and recent list
    const articleRes = await getArticleListApi({ page: 1, limit: 5 })
    stats.value.articleCount = articleRes.data.total
    recentArticles.value = articleRes.data.list

    // Sum views
    // In a real app, this should be a backend endpoint, but for now we mock or sum
    stats.value.totalViews = (articleRes.data.list || []).reduce(
      (acc: number, curr: any) => acc + (curr.views || 0),
      0,
    )

    // Fetch categories count
    const categoryRes = await getCategoryListApi()
    stats.value.categoryCount = categoryRes.data.list.length

    // Mock comment count
    stats.value.commentCount = 128
  } catch (error) {
    console.error(error)
  }
}

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.home {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 20px;
  .home-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 40px;
    margin-bottom: 20px;
    background-color: #f0f2f5;
    border-radius: 8px;
    .welcome-text {
      h2 {
        margin-bottom: 10px;
        font-size: 24px;
      }
      p {
        color: #666666;
      }
    }
    img {
      width: 200px;
    }
  }
  .statistics-box {
    margin-bottom: 20px;
    .statistic-item {
      display: flex;
      align-items: center;
      padding: 20px;
      background-color: #ffffff;
      border-radius: 8px;
      box-shadow: 0 2px 12px 0 rgb(0 0 0 / 5%);
      .item-left {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 60px;
        height: 60px;
        margin-right: 15px;
        background-color: #e6f7ff;
        border-radius: 50%;
        .el-icon {
          font-size: 30px;
          color: #1890ff;
        }
      }
      .item-right {
        .label {
          margin-bottom: 5px;
          font-size: 14px;
          color: #999999;
        }
        .value {
          font-size: 24px;
          font-weight: bold;
        }
      }
    }
  }
}
</style>
