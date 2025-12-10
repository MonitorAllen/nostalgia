<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import ArticleList from "@/components/article/ArticleList.vue";
import Breadcrumb from 'primevue/breadcrumb';

const route = useRoute();

// 从 URL 查询参数中获取 keyword (例如 /search?q=golang)
const keyword = computed(() => (route.query.q as string) || '');

// 面包屑配置
const home = { icon: 'pi pi-home', route: '/' };
const items = computed(() => [
  { label: '搜索结果' },
  { label: keyword.value } // 显示当前搜的词
]);
</script>

<template>
  <div class="flex flex-column w-full gap-3 px-0 md:px-2">
    <div class="surface-card shadow-1 border-round-md overflow-hidden">
      <Breadcrumb :home="home" :model="items" class="border-none p-3">
        <template #item="{ item }">
          <router-link v-if="item.route" :to="item.route" class="text-700 no-underline hover:text-green-500">
            <span :class="item.icon"></span>
            <span class="ml-2">{{ item.label }}</span>
          </router-link>
          <span v-else class="text-900 font-bold">
             {{ item.label }}
           </span>
        </template>
      </Breadcrumb>
    </div>

    <ArticleList :keyword="keyword" class="w-full"></ArticleList>
  </div>
</template>