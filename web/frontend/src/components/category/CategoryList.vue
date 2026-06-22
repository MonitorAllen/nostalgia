<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { listCategories } from '@/api/category'
import type { Category } from '@/types/category'
import { useToast } from '@/composables/useToast'
import AppBadge from '@/components/ui/AppBadge.vue'

const categories = ref<Category[]>([])
const toast = useToast()

const fetchCategories = async () => {
  try {
    const resp = await listCategories({ page: 1, limit: 20 })
    categories.value = resp.data.categories ?? []
  } catch (error: any) {
    toast.add({
      severity: 'error',
      summary: '获取分类失败',
      detail: error.response?.data?.error || '请稍后再试',
      life: 3000,
    })
  }
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
  </div>
</template>
