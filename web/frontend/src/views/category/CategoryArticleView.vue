<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getCategory } from '@/api/category'
import { useToast } from '@/composables/useToast'
import ArticleList from '@/components/article/ArticleList.vue'
import ArchiveTrail from '@/components/ui/ArchiveTrail.vue'
import { applySeoMetadata, buildCategorySeoMetadata } from '@/util/seo'

const props = defineProps({
  id: {
    type: Number,
    required: true,
  },
})

const route = useRoute()
const toast = useToast()
const categoryName = ref('分类')

const fetchCategory = async (id: number) => {
  try {
    const resp = await getCategory({ id })
    categoryName.value = resp.data.category.name
    applySeoMetadata(buildCategorySeoMetadata(id, categoryName.value))
  } catch (error: any) {
    toast.add({
      severity: 'error',
      summary: '获取分类失败',
      detail: error.response?.data?.message || '请稍后再试',
      life: 2500,
    })
  }
}

watch(
  () => route.params.id,
  (newId) => {
    const categoryId = Number(newId)
    if (!Number.isNaN(categoryId)) fetchCategory(categoryId)
  },
)

onMounted(async () => {
  await fetchCategory(props.id)
})
</script>

<template>
  <div class="space-y-4">
    <ArchiveTrail :items="[{ label: '分类索引' }, { label: categoryName }]" />
    <ArticleList :category-id="props.id" />
  </div>
</template>
