<script setup lang="ts">
import { ImagePlus, Trash2 } from '@lucide/vue'
import ArticleCover from '@/components/article/ArticleCover.vue'
import AppButton from '@/components/ui/AppButton.vue'
import type { ArticleCoverInspection } from '@/components/article/articleCoverPolicy'

withDefaults(
  defineProps<{
    cover?: string
    title?: string
    isUploading?: boolean
    inspection?: ArticleCoverInspection | null
  }>(),
  {
    cover: '',
    title: '',
    isUploading: false,
    inspection: null
  }
)

defineEmits<{
  upload: []
  remove: []
}>()
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center justify-between gap-3">
      <div class="min-w-0">
        <span class="block text-sm font-bold text-foreground">封面图</span>
        <span class="block text-xs leading-5 text-muted-foreground">
          推荐 1600x900，高清可用 1920x1080，最低建议 1200x675。
        </span>
      </div>
      <AppButton
        variant="secondary"
        size="sm"
        class="min-w-20 shrink-0 whitespace-nowrap"
        :disabled="isUploading"
        @click="$emit('upload')"
      >
        <ImagePlus class="size-4" aria-hidden="true" />
        {{ isUploading ? '上传中...' : '上传' }}
      </AppButton>
    </div>

    <div v-if="cover" class="space-y-3">
      <div
        v-if="inspection?.warnings?.length"
        class="rounded-archive border border-warning/45 bg-warning/10 p-3 text-xs font-semibold leading-5 text-warning"
      >
        <p v-for="warning in inspection.warnings" :key="warning" class="m-0">
          {{ warning }}
        </p>
      </div>

      <div class="rounded-archive border border-border bg-surface-raised p-3">
        <p class="m-0 mb-2 text-xs font-black text-muted-foreground">详情页头图</p>
        <ArticleCover :src="cover" :alt="title || '文章封面'" variant="preview" />
      </div>

      <div class="rounded-archive border border-border bg-surface-raised p-3">
        <p class="m-0 mb-2 text-xs font-black text-muted-foreground">列表卡片</p>
        <div class="grid overflow-hidden rounded-archive border border-border bg-surface md:grid-cols-[8rem_1fr]">
          <ArticleCover :src="cover" :alt="title || '文章封面'" variant="list" fallback-src="/images/go.png" />
          <div class="space-y-2 p-3">
            <div class="h-4 w-3/4 rounded bg-muted" />
            <div class="h-3 w-full rounded bg-muted" />
            <div class="h-3 w-2/3 rounded bg-muted" />
          </div>
        </div>
      </div>

      <div class="rounded-archive border border-border bg-surface-raised p-3">
        <p class="m-0 text-xs font-black text-muted-foreground">分享预览</p>
        <p class="m-0 mt-1 text-xs leading-5 text-muted-foreground">
          当前封面会作为 og:image 和 twitter:image 使用。
        </p>
      </div>

      <div class="flex justify-end">
        <AppButton variant="ghost" size="sm" class="text-danger hover:text-danger" @click="$emit('remove')">
          <Trash2 class="size-4" aria-hidden="true" />
          移除封面
        </AppButton>
      </div>
    </div>

    <div v-else class="rounded-archive border border-dashed border-border bg-surface-raised p-5 text-center">
      <p class="m-0 text-sm font-semibold text-muted-foreground">还没有封面图</p>
    </div>
  </div>
</template>
