<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { RouterLink } from 'vue-router';
import date from "@/util/date";
import { listArticle } from "@/api/article";
import type { Article } from "@/types/article";

// PrimeVue Components
import Paginator, { type PageState } from "primevue/paginator";
import { useToast } from "primevue/usetoast";
import Skeleton from 'primevue/skeleton';

const props = defineProps({
  categoryId: { type: Number, default: 0 }
})

const first = ref(0)
const currentPage = ref(1)
const limit = ref(10)
const totalRecords = ref(0)
const articles = ref<Article[]>([])
const loading = ref(false)
const toast = useToast()

const fetchArticles = async (categoryId: number, page: number, limit: number) => {
  loading.value = true
  try {
    const resp = await listArticle({categoryId, page, limit})
    totalRecords.value = resp.data.count
    articles.value = resp.data.articles
    if(page > 1) window.scrollTo({ top: 0, behavior: 'smooth' });
  } catch (err: any) {
    toast.add({ severity: 'error', summary: '错误', detail: err.response?.data?.error || '获取失败', life: 3000 })
  } finally {
    loading.value = false
  }
}

watch(() => props.categoryId, (newId) => {
  if (newId !== undefined) {
    first.value = 0
    currentPage.value = 1
    fetchArticles(newId, 1, limit.value)
  }
})

const onPageChange = (page: PageState) => {
  first.value = page.first
  currentPage.value = page.page + 1
  fetchArticles(props.categoryId, currentPage.value, limit.value)
}

const onImageError = (e: Event) => {
  (e.target as HTMLImageElement).src = '/images/go.png';
}

onMounted(() => {
  fetchArticles(props.categoryId, currentPage.value, limit.value)
})
</script>

<template>
  <div class="flex flex-column w-full relative px-0 pb-2">

    <div v-if="loading" class="flex flex-column gap-3">
      <div v-for="i in 3" :key="i" class="flex p-3 surface-card border-round-md h-10rem gap-3 shadow-1">
        <Skeleton width="30%" height="100%" class="border-round"></Skeleton>
        <div class="flex flex-column justify-content-between flex-1">
          <Skeleton width="60%" height="1.5rem" class="mb-2"></Skeleton>
          <Skeleton width="100%" height="3rem"></Skeleton>
          <div class="flex justify-content-between mt-2">
            <Skeleton width="4rem" height="1rem"></Skeleton>
            <Skeleton width="8rem" height="1rem"></Skeleton>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="totalRecords > 0" class="flex flex-column gap-3">
      <div v-for="(item, index) in articles" :key="item.id"
           class="flex flex-column sm:flex-row p-3 gap-3 border-round-md surface-card shadow-1 hover:shadow-4 transition-all transition-duration-300"
           :class="{ 'mt-0': index === 0 }">

        <div class="w-full sm:w-12rem flex-shrink-0 overflow-hidden border-round-md relative surface-ground">
          <RouterLink :to="`/article/${item.id}`" class="block h-12rem sm:h-10rem w-full">
            <img
                class="w-full h-full object-contain hover:scale-110 transition-transform transition-duration-500 bg-white"
                :src="item.cover || '/images/go.png'"
                @error="onImageError"
                alt="Cover"
            />
          </RouterLink>
        </div>

        <div class="flex flex-column flex-1 justify-content-between">
          <div class="flex flex-column gap-2">
            <RouterLink :to="`/article/${item.id}`" class="text-color no-underline transition-colors theme-hover-text">
              <h3 class="text-xl font-bold m-0 line-height-3">{{ item.title }}</h3>
            </RouterLink>
            <p class="text-color-secondary m-0 text-sm line-height-3 overflow-hidden text-overflow-ellipsis"
               style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;">
              {{ item.summary }}
            </p>
          </div>

          <div class="flex flex-wrap align-items-center justify-content-between mt-3 gap-2">
            <div class="flex align-items-center">
              <div class="theme-tag px-2 py-1 text-xs border-round font-medium flex align-items-center gap-1">
                <i class="pi pi-tag text-xs"></i>
                <span>{{ item.category_name }}</span>
              </div>
            </div>

            <div class="flex flex-wrap align-items-center gap-3 text-sm text-700 select-none">

              <div class="flex align-items-center gap-1">
                <i class="pi pi-user text-sm"></i>
                <span>{{ item.username }}</span>
              </div>

              <div class="hidden sm:flex align-items-center gap-1">
                <i class="pi pi-calendar text-sm"></i>
                <span>{{ date.format(item.created_at, 'YYYY-MM-DD') }}</span>
              </div>

              <div class="flex align-items-center gap-1">
                <i class="pi pi-heart text-sm"></i>
                <span>{{ item.likes }}</span>
              </div>

              <div class="flex align-items-center gap-1">
                <i class="pi pi-eye text-sm"></i>
                <span>{{ item.views }}</span>
              </div>

            </div>
          </div>
        </div>
      </div>

      <Paginator
          :first="first" :rows="limit" :totalRecords="totalRecords" @page="onPageChange"
          :template="{
                      '350px': 'FirstPageLink PrevPageLink NextPageLink LastPageLink',
                      default: 'FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink'
                    }"
          class="surface-card border-round-md shadow-1 mt-2"></Paginator>
    </div>

    <div v-else class="flex flex-column align-items-center justify-content-center py-6 surface-card border-round shadow-1">
      <i class="pi pi-inbox text-500 text-4xl mb-3"></i>
      <span class="text-500 font-medium">暂时没有文章</span>
    </div>

  </div>
</template>

<style scoped>
.object-cover {
  object-fit: cover;
}

/* 如果你极度介意图片被裁切，想看全图（但会有留白），
   可以将上面的 object-cover 改为 object-contain，并取消下面的注释 */
 .object-contain {
  object-fit: contain;
  background-color: #f8f9fa;
}

/* --- 主题色 #10b981 --- */
.theme-hover-text:hover {
  color: #10b981 !important;
}

.theme-tag {
  background-color: #ecfdf5;
  color: #10b981;
}
</style>