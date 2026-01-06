<script setup lang="ts">
import DataView from 'primevue/dataview';
import Badge from 'primevue/badge';

import {listCategories} from "@/api/category";
import {ref} from "vue";
import type {Category} from "@/types/category";
import {useToast} from "primevue/usetoast";

const first = ref(0)
const limit = ref(5)
const totalRecords = ref(0)
let categories = ref<Category[]>([])

const fetchCategories = async () => {
  try {
    const resp = await listCategories();
    categories.value = resp.data.categories
    totalRecords.value = resp.data.count
  } catch (error: any) {
    const toast = useToast();
    toast.add({
      severity: 'error',
      summary: '错误',
      detail: error.response?.data?.error || '获取分类列表失败',
      life: 3000
    })
  }
}

fetchCategories()

</script>

<template>
  <div class="flex flex-column justify-content-center">
    <DataView :value="categories" :first="first" :rows="limit" paginator
              paginator-template="PrevPageLink  NextPageLink">
      <template #list="slotProps">
        <div class="flex flex-column">
          <div v-for="(item, index) in slotProps.items" :key="index">
            <div class="flex flex-row justify-content-between align-items-center p-2"
                 :class="{ 'border-top-1 border-300 dark:border-surface-700': index !== 0}">
              <router-link :to="`/category/${item.id}`" class="cursor-pointer text-color hover:text-teal-400">
                <div>{{ item.name }}</div>
              </router-link>
              <Badge :value="item.article_count"></Badge>
            </div>
          </div>
        </div>
      </template>
    </DataView>
  </div>
</template>

<style scoped>
:deep(.p-paginator) {
  padding: 0 !important; /* 去掉容器内边距 */
  margin-top: 8px !important;
}

:deep(.p-paginator .p-paginator-prev),
:deep(.p-paginator .p-paginator-next),
:deep(.p-paginator .p-paginator-first),
:deep(.p-paginator .p-paginator-last) {
  height: 18px !important;
}
</style>