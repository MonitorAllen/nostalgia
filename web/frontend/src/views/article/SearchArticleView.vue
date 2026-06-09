<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import ArticleList from '@/components/article/ArticleList.vue'
import ArchiveTrail from '@/components/ui/ArchiveTrail.vue'

const route = useRoute()
const keyword = computed(() => (route.query.q as string) || '')
const items = computed(() => [{ label: '搜索结果' }, { label: keyword.value || '全部' }])
</script>

<template>
  <div class="space-y-4">
    <ArchiveTrail :items="items" />
    <ArticleList :keyword="keyword" />
  </div>
</template>
