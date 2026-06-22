<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ChevronLeft, ChevronRight } from '@lucide/vue'
import { listCategories } from '@/api/category'
import type { Category } from '@/types/category'
import { useToast } from '@/composables/useToast'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'

const categories = ref<Category[]>([])
const categoryPage = ref(1)
const categoryTotal = ref(0)
const categoryPageSize = 6
const toast = useToast()
const categoryTotalPages = computed(() => Math.max(1, Math.ceil(categoryTotal.value / categoryPageSize)))

const fetchCategories = async () => {
  try {
    const resp = await listCategories({ page: categoryPage.value, limit: categoryPageSize })
    categories.value = resp.data.categories ?? []
    categoryTotal.value = Number(resp.data.count || 0)
  } catch (error: any) {
    toast.add({
      severity: 'error',
      summary: '获取分类失败',
      detail: error.response?.data?.error || '请稍后再试',
      life: 3000,
    })
  }
}

const changeCategoryPage = (page: number) => {
  if (page < 1 || page > categoryTotalPages.value || page === categoryPage.value) return
  categoryPage.value = page
  void fetchCategories()
}

fetchCategories()
</script>

<template>
  <div class="space-y-2">
    <RouterLink
      v-for="item in categories"
      :key="item.id"
      :to="`/category/${item.id}`"
      class="flex items-center justify-between rounded-archive border border-transparent px-3 py-2 text-sm font-semibold transition hover:border-border hover:bg-muted/70"
    >
      <span>{{ item.name }}</span>
      <AppBadge>{{ item.article_count }}</AppBadge>
    </RouterLink>
    <p v-if="categories.length === 0" class="m-0 text-sm text-muted-foreground">
      暂无分类记录
    </p>
    <div v-if="categoryTotalPages > 1" class="mt-2 flex items-center justify-end gap-1">
      <AppButton
        variant="ghost"
        size="icon"
        class="size-8 text-muted-foreground"
        :disabled="categoryPage <= 1"
        aria-label="上一页"
        @click="changeCategoryPage(categoryPage - 1)"
      >
        <ChevronLeft class="size-4" aria-hidden="true" />
      </AppButton>
      <AppButton
        variant="ghost"
        size="icon"
        class="size-8 text-muted-foreground"
        :disabled="categoryPage >= categoryTotalPages"
        aria-label="下一页"
        @click="changeCategoryPage(categoryPage + 1)"
      >
        <ChevronRight class="size-4" aria-hidden="true" />
      </AppButton>
    </div>
  </div>
</template>
