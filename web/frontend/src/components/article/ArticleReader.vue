<script setup lang="ts">
import { computed } from 'vue'
import { Calendar, Clock, Eye, Heart } from '@lucide/vue'
import AppBadge from '@/components/ui/AppBadge.vue'
import date from '@/util/date'
import ArticleCover from './ArticleCover.vue'
import ArticleRichContent from './ArticleRichContent.vue'

type BadgeTone = 'neutral' | 'accent' | 'danger' | 'warning'

const props = withDefaults(
  defineProps<{
    title?: string
    summary?: string
    content?: string
    categoryName?: string
    readTime?: string
    cover?: string
    createdAt?: string
    likes?: number
    views?: number
    statusLabel?: string
    statusTone?: BadgeTone
    showMeta?: boolean
    showCover?: boolean
    previewContent?: boolean
  }>(),
  {
    title: '',
    summary: '',
    content: '',
    categoryName: '',
    readTime: '',
    cover: '',
    createdAt: '',
    likes: undefined,
    views: undefined,
    statusLabel: '',
    statusTone: 'neutral',
    showMeta: true,
    showCover: false,
    previewContent: false
  }
)

const displayTitle = computed(() => props.title?.trim() || '无标题文章')
const hasHeaderBadges = computed(
  () =>
    Boolean(props.categoryName?.trim()) ||
    Boolean(props.readTime?.trim()) ||
    Boolean(props.statusLabel?.trim())
)
const formattedCreatedAt = computed(() => {
  if (!props.createdAt) return ''
  const nextDate = new Date(props.createdAt)
  if (!Number.isFinite(nextDate.getTime())) return ''
  return date.format(props.createdAt, 'YYYY-MM-DD')
})
const showEngagementMeta = computed(
  () =>
    props.showMeta &&
    (Boolean(formattedCreatedAt.value) ||
      typeof props.likes === 'number' ||
      typeof props.views === 'number' ||
      Boolean(props.readTime?.trim()))
)
</script>

<template>
  <article class="archive-surface rounded-[1.1rem] p-5 md:p-8">
    <ArticleCover
      v-if="showCover && cover"
      :src="cover"
      :alt="displayTitle || '文章封面'"
      variant="detail"
      class="mb-6"
    />

    <header class="space-y-5 border-b border-border pb-6">
      <div v-if="hasHeaderBadges" class="flex flex-wrap gap-2">
        <AppBadge v-if="categoryName" tone="accent">{{ categoryName }}</AppBadge>
        <AppBadge v-if="readTime">{{ readTime }}</AppBadge>
        <AppBadge v-if="statusLabel" :tone="statusTone">{{ statusLabel }}</AppBadge>
      </div>

      <h1
        class="article-reader-title m-0 text-balance text-3xl font-extrabold leading-tight text-foreground md:text-4xl lg:text-[2.65rem]"
      >
        {{ displayTitle }}
      </h1>

      <div
        v-if="showEngagementMeta"
        class="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm font-semibold text-muted-foreground"
      >
        <span v-if="formattedCreatedAt" class="inline-flex items-center gap-1">
          <Calendar class="size-4" aria-hidden="true" />
          {{ formattedCreatedAt }}
        </span>
        <span v-if="typeof likes === 'number'" class="inline-flex items-center gap-1">
          <Heart class="size-4" aria-hidden="true" />
          {{ likes }}
        </span>
        <span v-if="typeof views === 'number'" class="inline-flex items-center gap-1">
          <Eye class="size-4" aria-hidden="true" />
          {{ views }}
        </span>
        <span v-if="readTime" class="inline-flex items-center gap-1">
          <Clock class="size-4" aria-hidden="true" />
          {{ readTime }}
        </span>
      </div>
    </header>

    <section v-if="summary" class="my-6 rounded-archive border border-border bg-surface-raised p-4">
      <p class="m-0 text-xs font-black uppercase text-muted-foreground">摘要</p>
      <p class="m-0 mt-2 text-pretty text-base leading-8 text-foreground/85">{{ summary }}</p>
    </section>

    <div class="article-reader-content mt-8">
      <ArticleRichContent :content="content" :preview="previewContent" />
    </div>
  </article>
</template>
