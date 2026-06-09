<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from '@lucide/vue'
import AppButton from './AppButton.vue'

const props = defineProps<{
  page: number
  rows: number
  totalRecords: number
}>()

const emit = defineEmits<{
  change: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.totalRecords / props.rows)))

const pages = computed(() => {
  const current = props.page
  const max = totalPages.value
  const start = Math.max(1, current - 2)
  const end = Math.min(max, current + 2)
  return Array.from({ length: end - start + 1 }, (_, index) => start + index)
})

const go = (nextPage: number) => {
  if (nextPage < 1 || nextPage > totalPages.value || nextPage === props.page) return
  emit('change', nextPage)
}
</script>

<template>
  <nav class="archive-glass flex items-center justify-between rounded-archive p-2" aria-label="分页">
    <AppButton variant="ghost" size="sm" :disabled="page <= 1" @click="go(page - 1)">
      <ChevronLeft class="h-4 w-4" />
      上一页
    </AppButton>
    <div class="hidden items-center gap-1 sm:flex">
      <button
        v-for="item in pages"
        :key="item"
        type="button"
        class="h-9 min-w-9 rounded-full px-3 text-sm font-semibold transition"
        :class="
          item === page
            ? 'bg-accent text-accent-foreground'
            : 'text-muted-foreground hover:bg-muted hover:text-foreground'
        "
        @click="go(item)"
      >
        {{ item }}
      </button>
    </div>
    <span class="text-sm font-semibold text-muted-foreground sm:hidden">
      {{ page }} / {{ totalPages }}
    </span>
    <AppButton variant="ghost" size="sm" :disabled="page >= totalPages" @click="go(page + 1)">
      下一页
      <ChevronRight class="h-4 w-4" />
    </AppButton>
  </nav>
</template>
