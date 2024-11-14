<template>
  <div class="flex flex-row justify-content-center">
    <div class="article-box flex flex-column m-auto p-2 mt-3 mb-3">
      <div class="flex flex-column align-items-center">
        <div>
          <h2 class="article-title">{{ title }}</h2>
        </div>
        <div class="flex flex-row gap-3">
          <div class="flex align-items-center">
            <i class="pi pi-calendar" style="font-size: .75rem"></i>
            <div class="font-medium text-xs ml-1">{{ date.format(create_at, 'YYYY-MM-DD HH:mm') }}</div>
          </div>
          <div class="flex align-items-center">
            <i class="pi pi-thumbs-up" style="font-size: .75rem"></i>
            <div class="font-medium text-xs ml-1">{{ likes }}</div>
          </div>
          <div class="flex align-items-center">
            <i class="pi pi-eye" style="font-size: .75rem; padding-top: 1px"></i>
            <div class="font-medium text-xs ml-1">{{ views }}</div>
          </div>
        </div>
      </div>
      <Divider />
      <div class="article-summary">
        <h3>简介</h3>
      </div>
      <div>
        {{summary}}
      </div>
      <Divider />
      <div class="flex flex-column" v-html="content"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import Divider from 'primevue/divider';

import Prism from 'prismjs';
import 'prismjs/components/prism-go.min.js';
import "prismjs/themes/prism-solarizedlight.css"

import { useArticleStore } from '@/store/module/article'
import date from '@/util/date'

const { id } = defineProps<{
  id?: string
}>()

const title = ref<string>('')
const summary = ref<string>('')
const content = ref<string>('')
const likes = ref<number>(0)
const views = ref<number>(0)
const create_at = ref<string>('')

const articleStore = useArticleStore()

onMounted(async () => {
  if (id != '') {
    await articleStore.getArticle(id as (string)).then((res: any) => {
      title.value = res.data.title
      summary.value = res.data.summary
      content.value = res.data.content
      likes.value = res.data.likes
      views.value = res.data.views
      create_at.value = res.data.create_at
      // 等待内容渲染完成后，激活语法高亮
    }).catch(err => console.log(err))

    Prism.highlightAll();
  }
})

</script>

<style scoped>
.article-box {
  width: 1000px;
  border: .1rem solid #ebebeb;
  background-color: rgb(239 239 239 / 0.3);
  .article-title {
    margin: 2px 0;
    height: fit-content;
    font-size: 26px;
  }

  .article-summary {
    border-left: 5px solid #20c997;
    background-color: #f7f7f7;
    font-size: 22px;
    margin-bottom: 5px;
    h3 {
      margin: 5px 10px;
    }
  }
}


</style>