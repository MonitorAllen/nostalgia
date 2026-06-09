<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { Calendar, Eye, Heart, Tag, User } from '@lucide/vue'
import date from '@/util/date'
import { listArticle, searchArticles } from '@/api/article'
import type { Article } from '@/types/article'
import TextHighlight from '@/components/common/TextHighlight.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import PaginationControl from '@/components/ui/PaginationControl.vue'
import AppBadge from '@/components/ui/AppBadge.vue'
import { useToast } from '@/composables/useToast'

const props = defineProps({
  categoryId: { type: Number, default: 0 },
  keyword: { type: String, default: '' },
})

const currentPage = ref(1)
const limit = ref(10)
const totalRecords = ref(0)
const articles = ref<Article[]>([])
const loading = ref(false)
const toast = useToast()

const fetchArticles = async (page: number, rows: number) => {
  loading.value = true
  try {
    const resp: any = props.keyword
      ? await searchArticles({ keyword: props.keyword, page, limit: rows })
      : await listArticle({ category_id: props.categoryId, page, limit: rows })

    totalRecords.value = resp.data.count
    articles.value = resp.data.articles ?? []
    if (page > 1) window.scrollTo({ top: 0, behavior: 'smooth' })
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: '获取文章失败',
      detail: err.response?.data?.error || '请稍后再试',
      life: 3000,
    })
  } finally {
    loading.value = false
  }
}

watch([() => props.categoryId, () => props.keyword], () => {
  currentPage.value = 1
  fetchArticles(currentPage.value, limit.value)
})

const onPageChange = (page: number) => {
  currentPage.value = page
  fetchArticles(currentPage.value, limit.value)
}

const onImageError = (e: Event) => {
  ;(e.target as HTMLImageElement).src = '/images/go.png'
}

onMounted(() => {
  fetchArticles(currentPage.value, limit.value)
})
</script>

<template>
  <section class="space-y-4">
    <template v-if="loading">
      <div v-for="i in 4" :key="i" class="archive-surface grid gap-4 rounded-archive p-4 md:grid-cols-[12rem_1fr]">
        <SkeletonBlock class="h-40 w-full md:h-32" />
        <div class="space-y-4">
          <SkeletonBlock class="h-6 w-2/3" />
          <SkeletonBlock class="h-16 w-full" />
          <SkeletonBlock class="h-5 w-1/2" />
        </div>
      </div>
    </template>

    <template v-else-if="totalRecords > 0">
      <article
        v-for="item in articles"
        :key="item.id"
        class="group archive-surface grid overflow-hidden rounded-archive transition duration-300 hover:-translate-y-0.5 hover:border-accent/35 md:grid-cols-[13rem_1fr]"
      >
        <RouterLink
          :to="`/article/${item.slug ? item.slug : item.id}`"
          class="relative block aspect-[16/10] overflow-hidden bg-muted md:aspect-auto md:h-full"
        >
          <img
            class="h-full w-full object-contain p-2 transition duration-500 group-hover:scale-[1.03]"
            :src="item.cover"
            :alt="item.title"
            @error="onImageError"
          />
        </RouterLink>

        <div class="flex min-w-0 flex-col justify-between gap-4 p-4">
          <div class="space-y-2">
            <div class="flex flex-wrap items-center gap-2">
              <AppBadge tone="accent">
                <Tag class="h-3.5 w-3.5" />
                {{ item.category_name }}
              </AppBadge>
              <span v-if="item.read_time" class="archive-label">{{ item.read_time }}</span>
            </div>

            <RouterLink :to="`/article/${item.slug ? item.slug : item.id}`" class="block">
              <h2 class="m-0 text-xl font-black leading-snug text-foreground transition group-hover:text-accent md:text-2xl">
                <TextHighlight :content="item.title" :keyword="props.keyword" />
              </h2>
            </RouterLink>

            <p
              class="m-0 text-sm leading-7 text-muted-foreground"
              style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;"
            >
              <TextHighlight :content="item.summary" :keyword="props.keyword" />
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs font-semibold text-muted-foreground">
            <span class="inline-flex items-center gap-1"><User class="h-3.5 w-3.5" />{{ item.username }}</span>
            <span class="inline-flex items-center gap-1"><Calendar class="h-3.5 w-3.5" />{{ date.format(item.created_at, 'YYYY-MM-DD') }}</span>
            <span class="inline-flex items-center gap-1"><Heart class="h-3.5 w-3.5" />{{ item.likes }}</span>
            <span class="inline-flex items-center gap-1"><Eye class="h-3.5 w-3.5" />{{ item.views }}</span>
          </div>
        </div>
      </article>

      <PaginationControl
        :page="currentPage"
        :rows="limit"
        :total-records="totalRecords"
        @change="onPageChange"
      />
    </template>

    <div v-else class="archive-surface rounded-archive p-10 text-center">
      <p class="m-0 text-lg font-bold text-foreground">还没有文章</p>
      <p class="m-0 mt-2 text-sm text-muted-foreground">暂时没有找到可展示的内容。</p>
    </div>
  </section>
</template>
