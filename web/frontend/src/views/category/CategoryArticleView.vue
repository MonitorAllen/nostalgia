<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { getCategory } from "@/api/category";
import { useToast } from "primevue/usetoast";

// Components
import ArticleList from "@/components/article/ArticleList.vue";
import Breadcrumb from 'primevue/breadcrumb';

const props = defineProps({
  id: {
    type: Number,
    required: true
  }
})

const route = useRoute();
const toast = useToast()

// 面包屑首页配置
const home = ref({
  icon: 'pi pi-home', // 使用图标看起来更简洁
  route: '/'
});

const items = ref<any[]>([]);

// 获取分类详情
const fetchCategory = async (id: number) => {
  try {
    const resp = await getCategory({ id: id })
    items.value = [{ label: resp.data.category.name }]
  } catch (error: any) {
    toast.add({
      severity: 'error',
      summary: '错误',
      detail: '获取分类失败: ' + error.response?.data?.message,
      life: 2500
    })
  }
}

// 监听路由参数变化（当从一个分类跳到另一个分类时）
watch(() => route.params.id, (newId) => {
  if (newId) {
    const categoryId = Number(newId)
    if (!isNaN(categoryId)) {
      fetchCategory(categoryId)
    }
  }
})

onMounted(async () => {
  await fetchCategory(props.id)
})
</script>

<template>
  <div class="flex flex-column w-full gap-3 px-0 md:px-2">

    <div class="surface-card shadow-1 border-round-md overflow-hidden">
      <Breadcrumb :home="home" :model="items" class="border-none p-3">
        <template #item="{ item }">
          <router-link
              v-if="item.route"
              :to="item.route"
              class="text-700 no-underline hover:text-emerald-500 transition-colors font-medium flex align-items-center gap-2"
          >
            <span v-if="item.icon" :class="item.icon"></span>
            <span>{{ item.label }}</span>
          </router-link>

          <span v-else class="text-900 font-bold flex align-items-center gap-2">
            <span v-if="item.icon" :class="item.icon"></span>
            <span>{{ item.label }}</span>
          </span>
        </template>
      </Breadcrumb>
    </div>

    <ArticleList :categoryId="props.id" class="w-full"></ArticleList>
  </div>
</template>

<style scoped>
.hover\:text-emerald-500:hover {
  color: #10b981 !important;
}
</style>