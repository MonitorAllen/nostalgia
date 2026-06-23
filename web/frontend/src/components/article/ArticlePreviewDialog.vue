<script setup lang="ts">
import { X } from '@lucide/vue'
import ArticleReader from './ArticleReader.vue'
import AppButton from '@/components/ui/AppButton.vue'

type BadgeTone = 'neutral' | 'accent' | 'danger' | 'warning'

withDefaults(
  defineProps<{
    open: boolean
    loading?: boolean
    dialogTitle?: string
    eyebrow?: string
    articleTitle?: string
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
  }>(),
  {
    loading: false,
    dialogTitle: '文章预览',
    eyebrow: '实际阅读效果',
    articleTitle: '',
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
    showMeta: false,
    showCover: false
  }
)

defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 overflow-y-auto bg-background/72 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-labelledby="article-preview-dialog-title"
      @click.self="$emit('close')"
    >
      <div class="mx-auto my-6 w-full max-w-[820px]">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div>
            <p class="m-0 text-xs font-bold text-muted-foreground">{{ eyebrow }}</p>
            <h2 id="article-preview-dialog-title" class="m-0 text-xl font-black text-foreground">
              {{ dialogTitle }}
            </h2>
          </div>
          <AppButton variant="secondary" size="icon" aria-label="关闭预览" @click="$emit('close')">
            <X class="size-4" aria-hidden="true" />
          </AppButton>
        </div>

        <section
          v-if="loading"
          class="archive-surface rounded-[1.1rem] p-8 text-center text-sm font-semibold text-muted-foreground"
          aria-live="polite"
        >
          正在读取文章预览
        </section>

        <ArticleReader
          v-else
          :title="articleTitle"
          :summary="summary"
          :content="content"
          :category-name="categoryName"
          :read-time="readTime"
          :cover="cover"
          :created-at="createdAt"
          :likes="likes"
          :views="views"
          :status-label="statusLabel"
          :status-tone="statusTone"
          :show-meta="showMeta"
          :show-cover="showCover"
          preview-content
        />
      </div>
    </div>
  </Teleport>
</template>
