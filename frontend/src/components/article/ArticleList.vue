<script setup lang="ts">

import date from "@/util/date";
import Paginator, {type PageState} from "primevue/paginator";
import ProgressSpinner from 'primevue/progressspinner';
import {useToast} from "primevue/usetoast";
import {onMounted, ref, watch} from "vue";
import type {Article} from "@/types/article";
import {listArticle} from "@/api/article";

const props = defineProps({
  categoryId: {
    type: Number,
    default: 0
  }
})


const first = ref(0)
const currentPage = ref(1)
const limit = ref(10)
const totalRecords = ref(0)
const articles = ref<Article[]>([])

// 添加加载状态
const loading = ref(false)
const error = ref('')

const fetchArticles = async (categoryId: number, page: number, limit: number) => {
  loading.value = true
  error.value = ''
  try {
    const resp = await listArticle({categoryId, page, limit})
    totalRecords.value = resp.data.count
    articles.value = resp.data.articles
  } catch (err: any) {
    const toast = useToast()
    error.value = err.response?.data?.error || '获取文章列表失败'
    toast.add({
      severity: 'error',
      summary: '错误',
      detail: error.value,
      life: 3000
    })
  } finally {
    loading.value = false
  }
}

// 监听 categoryId 变化
watch(() => props.categoryId, (newId: number, oldId) => {
  if (newId) {
    first.value = 0
    currentPage.value = 1
    fetchArticles(newId, currentPage.value, limit.value)
  }
})

const onPageChange = (page: PageState) => {
  first.value = page.first
  currentPage.value = page.page + 1
  fetchArticles(props.categoryId, page.page + 1, limit.value)
}

onMounted(() => {
  fetchArticles(props.categoryId, currentPage.value, limit.value)
})
</script>

<template>
  <div class="flex w-full relative px-2 center">
    <!-- 加载状态 -->
    <div v-if="loading" class="flex justify-content-center mt-4">
      <ProgressSpinner size="50" />
    </div>

    <div v-else-if="totalRecords > 0" class="w-full">
        <div class="flex flex-row w-full mt-2 p-3 gap-3 transition-all transition-duration-500 article-box border-round-md"
             :class="{ 'mt-0': index === 0 }"
             v-for="(item, index) in articles"
             :key="index">
          <div class="flex flex-row w-3 align-items-center article-cover">
            <a :herf="`/article/${item.id}`">
              <img
                  class="w-full "
                  src="/images/go.png"
                  alt="Image"
              />
            </a>
          </div>
          <div class="flex flex-column w-9 justify-content-between">
            <div class="flex flex-column justify-content-start">
              <div class="flex font-medium text-2xl text-primary">
                <a class="text-green-500" :href="`/article/${item.id}`" target="_blank">{{ item.title }}
                </a>
              </div>
              <div class="text-base">
                <p>
                  {{ item.summary.length > 100 ? item.summary.substring(0, 100) + "……" : item.summary }}
                </p>
              </div>
            </div>
            <div class="flex flex-wrap sm:flex-row gap-3 justify-content-between mt-3 md:mt-0 lg:mt-3">
              <div class="flex flex-row gap-3">
                <div class="flex align-items-center">
                  <i class="pi pi-tag" style="font-size: .75rem;"></i>
                  <div class="font-medium text-xs ml-1 ">{{ item.category_name }}</div>
                </div>
              </div>
              <div class="flex flex-wrap sm:flex-row gap-3">
                <div class="flex align-items-center">
                  <i class="pi pi-user" style="font-size: .75rem"></i>
                  <div class="font-medium text-xs ml-1 ">{{ item.username }}</div>
                </div>
                <div class="flex align-items-center">
                  <i class="pi pi-calendar" style="font-size: .75rem"></i>
                  <div class="font-medium text-xs ml-1">{{ date.format(item.created_at, 'YYYY-MM-DD') }}</div>
                </div>
                <div class="flex align-items-center">
                  <i class="pi pi-heart" style="font-size: .75rem"></i>
                  <div class="font-medium text-xs ml-1">{{ item.likes }}</div>
                </div>
                <div class="flex align-items-center">
                  <i class="pi pi-eye" style="font-size: .75rem; padding-top: 1px"></i>
                  <div class="font-medium text-xs ml-1">{{ item.views }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      <Paginator :first="first" :rows="limit" :totalRecords="totalRecords" @page="onPageChange"></Paginator>
    </div>
    <div
        class="flex w-full align-items-center justify-content-center mt-2 h-7rem font-bold border-dashed border-round border-300"
        v-else>
      <span class="text-color-secondary" style="letter-spacing: 8px;">这个家伙很懒什么都没有留下</span>
    </div>
  </div>
</template>

<style scoped>
.center {
  .article-box {
    background-color: #fbfbfb;

    .article-cover {
      img {
        vertical-align: middle;
        border-style: none;
      }
    }
  }
}

.full-width {
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