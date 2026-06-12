<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { CalendarDays, Eye, FileText, Folder, Heart, Pencil, Plus, Search, Trash2 } from '@lucide/vue'
import { isAutomationDraft } from '@/admin/articleAutomation'
import type { AdminArticle } from '@/admin/types'
import {
  deleteAdminArticle,
  listAdminArticles,
  updateAdminArticle,
} from '@/admin/api/adminArticleApi'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const toast = useToast()

const articles = ref<AdminArticle[]>([])
const loading = ref(false)
const total = ref(0)
const selectedArticle = ref<AdminArticle | null>(null)
const deleting = ref(false)
const activeAction = ref('')

const query = reactive({
  title: '',
  page: 1,
  limit: 12,
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / query.limit)))
const hasArticles = computed(() => articles.value.length > 0)
const showingFrom = computed(() => (total.value === 0 ? 0 : (query.page - 1) * query.limit + 1))
const showingTo = computed(() => Math.min(query.page * query.limit, total.value))

const fetchArticles = async () => {
  if (loading.value) return

  loading.value = true

  try {
    const response = await listAdminArticles({
      title: query.title.trim() || undefined,
      page: query.page,
      limit: query.limit,
    })

    articles.value = response.data.articles ?? []
    total.value = Number(response.data.count || 0)
  } catch {
    // Admin HTTP client already shows a toast for request failures.
  } finally {
    loading.value = false
  }
}

const searchArticles = () => {
  query.page = 1
  void fetchArticles()
}

const clearSearch = () => {
  query.title = ''
  searchArticles()
}

const createArticle = () => {
  void router.push({ name: 'adminArticleNew' })
}

const editArticle = (id: string) => {
  void router.push({ name: 'adminArticleEdit', params: { id } })
}

const togglePublish = async (article: AdminArticle) => {
  const nextStatus = !article.is_publish
  activeAction.value = `publish:${article.id}`

  try {
    await updateAdminArticle({ id: article.id, is_publish: nextStatus })
    article.is_publish = nextStatus
    toast.add({
      severity: 'success',
      summary: nextStatus ? '文章已发布' : '文章已转为草稿',
      detail: article.title || '无标题文章',
      life: 2400,
    })
  } catch {
    // Admin HTTP client already shows a toast for request failures.
  } finally {
    activeAction.value = ''
  }
}

const askDelete = (article: AdminArticle) => {
  selectedArticle.value = article
}

const cancelDelete = () => {
  if (deleting.value) return
  selectedArticle.value = null
}

const confirmDelete = async () => {
  if (!selectedArticle.value || deleting.value) return

  const article = selectedArticle.value
  deleting.value = true

  try {
    await deleteAdminArticle(article.id)
    articles.value = articles.value.filter((item) => item.id !== article.id)
    total.value = Math.max(0, total.value - 1)
    selectedArticle.value = null
    toast.add({
      severity: 'success',
      summary: '文章已删除',
      detail: article.title || '无标题文章',
      life: 2400,
    })

    if (articles.value.length === 0 && query.page > 1) {
      query.page -= 1
      await fetchArticles()
    }
  } catch {
    // Admin HTTP client already shows a toast for request failures.
  } finally {
    deleting.value = false
  }
}

const changePage = (page: number) => {
  if (page < 1 || page > totalPages.value || page === query.page) return
  query.page = page
  void fetchArticles()
}

const formatDate = (value?: string) => {
  if (!value) return '未记录'

  const date = new Date(value)
  if (!Number.isFinite(date.getTime())) return '未记录'

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const numberLabel = (value: number) => {
  return new Intl.NumberFormat('zh-CN').format(value || 0)
}

const summaryLabel = (article: AdminArticle) => {
  return article.summary?.trim() || '暂无摘要'
}

const categoryLabel = (article: AdminArticle) => {
  return article.category_name?.trim() || '未分类'
}

const isActionBusy = (key: string) => activeAction.value === key

onMounted(() => {
  void fetchArticles()
})
</script>

<template>
  <main class="space-y-5">
    <header class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div class="min-w-0 space-y-2">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="m-0 text-2xl font-black leading-tight text-foreground text-balance">
            文章管理
          </h1>
          <AppBadge tone="neutral" class="tabular-nums">共 {{ numberLabel(total) }} 篇</AppBadge>
        </div>
        <p class="m-0 max-w-2xl text-sm leading-6 text-muted-foreground text-pretty">
          快速检查文章状态、调整发布节奏，或者进入编辑器继续打磨内容。
        </p>
      </div>

      <div class="flex w-full flex-col gap-3 sm:flex-row lg:w-auto">
        <form class="flex min-w-0 flex-1 gap-2 lg:w-80" role="search" @submit.prevent="searchArticles">
          <label class="sr-only" for="admin-article-search">搜索文章标题</label>
          <AppInput
            id="admin-article-search"
            v-model="query.title"
            placeholder="搜索文章标题"
            class="min-w-0"
          />
          <AppButton type="submit" variant="secondary" :disabled="loading">
            <Search class="size-4" aria-hidden="true" />
            搜索
          </AppButton>
        </form>

        <AppButton class="sm:shrink-0" @click="createArticle">
          <Plus class="size-4" aria-hidden="true" />
          写新文章
        </AppButton>
      </div>
    </header>

    <section
      v-if="loading"
      class="archive-surface rounded-archive p-8 text-center text-sm font-semibold text-muted-foreground"
      aria-live="polite"
    >
      正在读取文章
    </section>

    <section v-else-if="hasArticles" class="space-y-3" aria-label="文章列表">
      <article
        v-for="article in articles"
        :key="article.id"
        class="archive-surface rounded-archive p-4 transition duration-200 hover:border-accent/35 hover:bg-surface-raised/70"
      >
        <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
          <div class="min-w-0 flex-1 space-y-3">
            <div class="flex flex-wrap items-center gap-2">
              <AppBadge :tone="article.is_publish ? 'accent' : 'neutral'">
                {{ article.is_publish ? '已发布' : '草稿' }}
              </AppBadge>
              <AppBadge v-if="isAutomationDraft(article)" tone="warning">
                自动化草稿
              </AppBadge>
              <AppBadge tone="neutral">
                <Folder class="size-3.5" aria-hidden="true" />
                {{ categoryLabel(article) }}
              </AppBadge>
            </div>

            <button
              type="button"
              class="block max-w-full text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
              @click="editArticle(article.id)"
            >
              <h2
                class="m-0 truncate text-lg font-black leading-snug text-foreground transition-colors hover:text-accent"
                :title="article.title || '无标题文章'"
              >
                {{ article.title || '无标题文章' }}
              </h2>
            </button>

            <p
              class="m-0 max-w-3xl text-sm leading-6 text-muted-foreground text-pretty"
              style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;"
            >
              {{ summaryLabel(article) }}
            </p>

            <dl class="m-0 flex flex-wrap gap-x-4 gap-y-2 text-xs font-semibold text-muted-foreground">
              <div class="flex items-center gap-1.5">
                <dt class="sr-only">更新时间</dt>
                <CalendarDays class="size-3.5" aria-hidden="true" />
                <dd class="m-0 tabular-nums">{{ formatDate(article.updated_at || article.created_at) }}</dd>
              </div>
              <div class="flex items-center gap-1.5">
                <dt class="sr-only">浏览</dt>
                <Eye class="size-3.5" aria-hidden="true" />
                <dd class="m-0 tabular-nums">{{ numberLabel(article.views) }}</dd>
              </div>
              <div class="flex items-center gap-1.5">
                <dt class="sr-only">喜欢</dt>
                <Heart class="size-3.5" aria-hidden="true" />
                <dd class="m-0 tabular-nums">{{ numberLabel(article.likes) }}</dd>
              </div>
              <div class="flex items-center gap-1.5">
                <dt class="sr-only">文章编号</dt>
                <FileText class="size-3.5" aria-hidden="true" />
                <dd class="m-0 max-w-[14rem] truncate font-mono text-[0.72rem]">{{ article.id }}</dd>
              </div>
            </dl>
          </div>

          <div class="flex flex-wrap items-center gap-2 xl:justify-end">
            <AppButton variant="secondary" size="sm" @click="editArticle(article.id)">
              <Pencil class="size-4" aria-hidden="true" />
              编辑
            </AppButton>
            <AppButton
              variant="ghost"
              size="sm"
              :disabled="isActionBusy(`publish:${article.id}`)"
              @click="togglePublish(article)"
            >
              {{ article.is_publish ? '设为草稿' : '发布文章' }}
            </AppButton>
            <AppButton variant="ghost" size="sm" class="text-danger hover:text-danger" @click="askDelete(article)">
              <Trash2 class="size-4" aria-hidden="true" />
              删除
            </AppButton>
          </div>
        </div>
      </article>
    </section>

    <section v-else class="archive-surface rounded-archive p-8 text-center">
      <p class="m-0 text-lg font-black text-foreground">还没有匹配的文章</p>
      <p class="m-0 mt-2 text-sm leading-6 text-muted-foreground">
        {{ query.title ? '换个关键词再试，或清空搜索继续浏览全部文章。' : '开始写第一篇文章，后台列表会在这里记录它的发布状态。' }}
      </p>
      <div class="mt-5 flex flex-wrap justify-center gap-2">
        <AppButton v-if="query.title" variant="secondary" @click="clearSearch">清空搜索</AppButton>
        <AppButton @click="createArticle">
          <Plus class="size-4" aria-hidden="true" />
          写新文章
        </AppButton>
      </div>
    </section>

    <nav
      v-if="!loading && total > query.limit"
      class="archive-glass flex flex-col gap-3 rounded-archive p-3 sm:flex-row sm:items-center sm:justify-between"
      aria-label="文章分页"
    >
      <p class="m-0 text-sm font-semibold text-muted-foreground">
        <span class="tabular-nums">{{ showingFrom }}-{{ showingTo }}</span>
        /
        <span class="tabular-nums">{{ numberLabel(total) }}</span>
      </p>
      <div class="flex items-center justify-between gap-2 sm:justify-end">
        <AppButton variant="ghost" size="sm" :disabled="query.page <= 1" @click="changePage(query.page - 1)">
          上一页
        </AppButton>
        <span class="min-w-20 text-center text-sm font-semibold text-muted-foreground tabular-nums">
          {{ query.page }} / {{ totalPages }}
        </span>
        <AppButton
          variant="ghost"
          size="sm"
          :disabled="query.page >= totalPages"
          @click="changePage(query.page + 1)"
        >
          下一页
        </AppButton>
      </div>
    </nav>

    <ConfirmDialog
      :open="Boolean(selectedArticle)"
      title="删除文章"
      :description="`确认删除「${selectedArticle?.title || '无标题文章'}」吗？这个操作不可恢复。`"
      :confirm-label="deleting ? '删除中...' : '删除文章'"
      cancel-label="取消"
      danger
      @cancel="cancelDelete"
      @confirm="confirmDelete"
    />
  </main>
</template>
