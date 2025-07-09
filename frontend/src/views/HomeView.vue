<template>
  <div class="w-full px-3 mx-auto main">
    <div class="flex flex-wrap mt-3 bg-white">
      <div class="block w-full relative px-2 center">
        <div v-if="totalRecords > 0">
          <div class="flex flex-column">
            <div class="flex flex-row w-full mt-2 p-3 gap-3 transition-all transition-duration-500 article-box"
                 :class="{ 'mt-0': index === 0 }"
                 v-for="(item, index) in articles"
                 :key="index">
              <div class="flex flex-row w-3 align-items-center article-cover">
                <img
                  class="w-full max-h-10rem"
                  src="/images/go.png"
                  alt="Image"
                />
              </div>
              <div class="flex flex-column w-9 justify-content-between">
                <div class="flex flex-column justify-content-start">
                  <div class="flex text-base font-medium text-2xl text-primary">
                    <a class="text-green-300" :href="`/article/${item.id}`" target="_blank">{{ item.title }}
                    </a>
                  </div>
                  <div class="mt-2 text-sm">
                    <p>
                      {{ item.summary.length > 100 ? item.summary.substring(0, 100) + "……": item.summary }}
                    </p>
                  </div>
                </div>
                <div class="flex flex-row gap-3 justify-content-between min-w-min">
                  <div class="flex flex-row gap-3">
                    <div class="flex align-items-center" v-if="item.tags.length > 0">
                      <Tag class="ml-1 p-1" severity="success" style="font-size: .75rem; font-family: 等线,serif"
                           :key="index" :value="tag" v-for="(tag, index) in item.tags"></Tag>
                    </div>
                  </div>
                  <div class="flex flex-row gap-3 line-height-1">
                    <div class="flex">
                      <i class="pi pi-user" style="font-size: .75rem"></i>
                      <div class="font-medium text-xs ml-1 ">{{ item.username }}</div>
                    </div>
                    <div class="flex align-items-center">
                      <i class="pi pi-calendar" style="font-size: .75rem"></i>
                      <div class="font-medium text-xs ml-1">{{ date.format(item.created_at, 'YYYY-MM-DD') }}</div>
                    </div>
                    <div class="flex align-items-center">
                      <i class="pi pi-thumbs-up" style="font-size: .75rem"></i>
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
          </div>
          <Paginator :rows="limit" :totalRecords="totalRecords" @page="onPageChange"></Paginator>
        </div>
        <div class="flex align-items-center justify-content-center mt-2 h-7rem font-bold border-dashed border-round border-300" v-else>
          <span class="text-color-secondary ">这个家伙很懒什么都没有留下</span>
        </div>
      </div>
      <div class="block w-full relative px-2 mt-2 left-side">
        <div class="flex flex-column border-round-md shadow-1 mb-3 todo-list">
          <div class="flex w-full border-round-top-md justify-content-start align-items-center gap-2 left-side-menu-title">
            <span class="pi pi-clipboard ml-1"></span>
            <span>代办事项</span>
          </div>
          <div id="todo-list-content" class="flex flex-column row-gap-2 p-2">
            <div class="flex flex-row gap-2 align-items-center">
              <Checkbox binary />
              <label>完善CI/CD</label>
            </div>
          </div>
        </div>
        <div class="flex flex-column border-round-md shadow-1 mb-3">
          <div class="flex border-round-top-md justify-content-start align-items-center gap-2 left-side-menu-title">
            <span class="pi pi-github ml-1"></span>
            <span>近期活动</span>
          </div>
          <div class="p-2">
            <GithubContributions></GithubContributions>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Checkbox from 'primevue/checkbox';
import Tag from 'primevue/tag'
import Paginator, { type PageState } from 'primevue/paginator'

import { useToast } from 'primevue/usetoast'

import { useArticleStore } from '@/store/module/article'
import { ref, onMounted } from 'vue'

import date from '@/util/date'
import type { Article } from '@/types/article'
import GithubContributions from "@/components/GithubContributions.vue";

const toast = useToast()

const first = ref(0)
const currentPage = ref(1)
const limit = ref(10)
const totalRecords = ref(0)
const articles = ref<Article[]>([])

const articleStore = useArticleStore()

const onPageChange = (page: PageState) => {
  first.value = 0
  console.log(page)
  currentPage.value = page.page
  articleStore.listArticle(page.page + 1, limit.value)
    .then((res: any) => {
      if (res.data.articles.length > 0) {
        totalRecords.value = res.data.count
        articles.value = res.data.articles
      }
    }).catch((err) => {
    toast.add({ severity: 'error', summary: 'Error Message', detail: err.response.data.error })
  })
}

onMounted(() => {
  articleStore.listArticle(currentPage.value, limit.value)
    .then((res: any) => {
      if (res.data.articles.length > 0) {
        totalRecords.value = res.data.count
        articles.value = res.data.articles
      }
    }).catch((err) => {
    toast.add({ severity: 'error', summary: 'Error Message', detail: err.response.data.error })
  })
})
</script>

<style scoped>
.main {
  min-width: 330px;
}

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

.left-side-menu-title {
  height: 2rem !important;
  background-color: var(--green-200);
}

@media (min-width: 576px) {
  .main {
    max-width: 540px;
  }
}

@media (min-width: 768px) {
  .main {
    max-width: 720px;
  }
}

@media (min-width: 992px) {
  .main {
    max-width: 960px;
  }

  .center {
    flex: 0 0 66.666667%;
    max-width: 66.666667%;
  }

  .left-side {
    flex: 0 0 33.333333%;
    max-width: 33.333333%;
  }
}

@media (min-width: 1200px) {
  .main {
    max-width: 1140px;
  }
}
</style>
