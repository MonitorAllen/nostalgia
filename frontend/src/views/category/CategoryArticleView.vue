<script setup lang="ts">
import {onMounted, ref, watch} from "vue";
import ArticleList from "@/components/article/ArticleList.vue";
import {getCategory} from "@/api/category";
import {useToast} from "primevue/usetoast";

import Breadcrumb from 'primevue/breadcrumb';
import {useRoute} from "vue-router";

const props = defineProps({
  id: {
    type: Number,
    required: true
  }
})

const home = ref({
  label: '首页',
  route: '/'
});
let items: any = ref(null);

const toast = useToast()

const route = useRoute();

watch(() => route.params.id, (newId) => {
  if (newId) {
    const categoryId = Number(newId)
    if (!isNaN(categoryId)) {
      fetchCategory(categoryId)
    }
  }
})

const fetchCategory = async (id: number) => {
  try {
    const resp = await getCategory({id: id})
    items.value = [{label: resp.data.category.name}]
  } catch (error: any) {
    toast.add({
      severity: 'error',
      summary: '错误',
      detail: '获取分类失败: ' + error.response?.data?.message,
      life: 2500
    })
  }
}

onMounted(async () => {
  await fetchCategory(props.id)
})
</script>

<template>
  <div class="block w-full relative center">
    <div class="px-2 mt-2">
      <Breadcrumb :home="home" :model="items" class="bg-gray-50 w-full border-round-md">
        <template #item="{ item }">
          <router-link v-if="item.route" :to="item.route" class="text-color cursor-pointer">
              <span>{{ item.label }}</span>
          </router-link>
          <span v-else class="text-700">{{ item.label }}</span>
        </template>
      </Breadcrumb>
    </div>
    <ArticleList :categoryId="props.id" class="mt-0 full-width"></ArticleList>
  </div>
</template>

<style scoped>
.center {
  flex: 0 0 100%;
  max-width: 100%;
}

/* 覆盖 ArticleList 内部的 center 类 */
.full-width :deep(.center) {
  flex: none !important;
  max-width: none !important;
  width: 100% !important;
}

@media (min-width: 992px) {
  .center {
    flex: 0 0 66.666667%;
    max-width: 66.666667%;
  }
}
</style>