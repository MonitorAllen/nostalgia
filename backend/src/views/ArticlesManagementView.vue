<template>
    <div class="surface-card h-full shadow-2 border-round p-3">
      <div class="flex justify-content-between align-items-center mb-4">
        <h1 class="text-xl font-medium m-0">文章管理</h1>
        <router-link :to="{name: 'editor'}" target="_blank">
          <Button label="新建文章" icon="pi pi-plus" severity="primary"/>
        </router-link>
      </div>
      <div style="height: 580px;">
        <DataTable
          :value="articles"
          scrollable
          scrollHeight="580px"
          showGridlines
          stripedRows
          responsiveLayout="scroll"
          class="p-datatable-sm mb-4"
        >
          <Column field="title" header="标题" style="min-width: 140px;"></Column>
          <Column field="summary" header="摘要" style="min-width: 120px;"></Column>
          <Column field="likes" header="点赞" style="width: 50px; min-width: 50px;"></Column>
          <Column field="views" header="浏览量" style="width: 65px; min-width: 65px;"></Column>
          <Column field="is_publish" header="是否发布" style="width: 80px; min-width: 80px;">
            <template #body="slotProps">
              <Tag :value="slotProps.data.is_publish ? '已发布' : '未发布'" :severity="getStatusSeverity(slotProps.data.is_publish)" />
            </template>
          </Column>
          <Column field="created_at" header="创建时间" style="width: 172px; min-width: 172px">
            <template #body="slotProps">
              {{ format(slotProps.data.created_at, 'YYYY-MM-DD HH:mm:ss') }}
            </template>
          </Column>
          <Column field="updated_at" header="更新时间" style="width: 172px; min-width: 172px;">
            <template #body="slotProps">
              {{ format(slotProps.data.updated_at, 'YYYY-MM-DD HH:mm:ss') }}
            </template>
          </Column>
          <Column header="操作" :exportable="false" style="width: 15rem; min-width:15rem">
            <template #body="slotProps">
              <Button label="编辑信息" size="small" severity="info" class="mr-2" @click="editArticle(slotProps.index, slotProps.data.id)"/>
              <router-link :to="{name: 'editor', params: { id: slotProps.data.id }}" target="_blank" class="mr-2">
                <Button label="编辑内容" size="small" severity="info" />
              </router-link>
              <Button label="删除" size="small" severity="danger" />
            </template>
          </Column>
        </DataTable>
      </div>
      <Paginator
        :rows="limit"
        :totalRecords="count"
        :first="first"
        template="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
        :rowsPerPageOptions="[10,20,50]"
        currentPageReportTemplate="显示第 {first} 到 {last} 条记录，共 {totalRecords} 条"
        @page="onPage"
      />
    </div>

    <Dialog v-model:visible="visible" modal header="编辑文章信息" :style="{ width: '45rem' }">
      <Form v-slot="$form" :initialValues="article" @submit="onUpdateArticle" class="flex flex-column">
        <InputText type="hidden" name="id"></InputText>
        <div class="flex items-center gap-4 mb-4">
            <label for="title" class="font-semibold w-24">标题</label>
            <InputText id="title" name="title" class="flex-auto"/>
        </div>
        <div class="flex felx-row gap-4 mb-4">
            <label for="summary" class="font-semibold w-24">摘要</label>
            <Textarea id="summary" name="summary" autoResize class="flex-auto"/>
        </div>
        <div class="flex gap-4 mb-8">
            <label for="is_publish" class="font-semibold w-24">发布</label>
            <ToggleSwitch id="is_publish" name="is_publish"/>
        </div>
        <div class="flex justify-content-end gap-2">
            <Button type="button" label="取消" severity="secondary" @click="visible = false"></Button>
            <Button type="submit" label="保存"></Button>
        </div>
      </Form>
    </Dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue' 
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Paginator from 'primevue/paginator'
import Dialog from 'primevue/dialog';
import InputText from 'primevue/inputtext';
import Textarea from 'primevue/textarea';
import ToggleSwitch from 'primevue/toggleswitch';
import { Form } from '@primevue/forms';
import { useArticleStore, type Article } from '@/stores/article'
import { useToast } from 'primevue/usetoast'
import format from '@/util/date'
import { fetchArticleById, type UpdateArticleParams, updateArticle } from '@/api/articles'

const articleStore = useArticleStore()
const toast = useToast()

const articles = ref<Article[]>([])
const limit = ref(10)
const count = ref(0)
const first = ref(0)

onMounted(async () => {
    try {
        await articleStore.listAllArticles(1, limit.value)
        articles.value = articleStore.articles
        count.value = articleStore.count
    } catch (error) {
        console.log(error)
        toast.add({ severity: 'error', summary: '错误', detail: '获取文章列表失败', life: 2500})
    }
})

const onPage = async (event: any) => {
    first.value = event.first
    const page = Math.floor(event.first / event.rows) + 1
    limit.value = event.rows
    await articleStore.listAllArticles(page, limit.value)
    articles.value = articleStore.articles
    count.value = articleStore.count
}

const getStatusSeverity = (status: boolean) => {
  switch (status) {
    case true:
      return 'success'
    case false:
      return 'danger'
    default:
      return 'info'
  }
}

type EditArticle = {
  id: string
  title: string
  summary: string
  is_publish: boolean
}

const visible = ref(false)
let article = reactive<Article>({
  id: '',
  title: '',
  summary: '',
  is_publish: false,
  likes: 0,
  views: 0,
  created_at: '',
  updated_at: '',
  owner: ''
})

const edittingIndex = ref(-1)

const editArticle = async (index: number, id: string) => {
  try {
    const res = await fetchArticleById(id, false)
    Object.assign(article, res)
    visible.value = true
    edittingIndex.value = index
  } catch (error: any) {
    console.log(error)
    toast.add({severity: 'error', summary: '错误', detail: '获取文章信息失败: ' + error.response?.data?.message, life: 2500})
  }
}

const onUpdateArticle = async (e: any) => {
  try {
    const res = await updateArticle(e.values as UpdateArticleParams)
    if (edittingIndex.value != -1) {
        Object.assign(articles.value[edittingIndex.value], res)
        toast.add({severity: 'success', summary: '成功', detail: '保存文章成功', life: 2500})
    }
  } catch (error: any) {
    toast.add({severity: 'error', summary: '错误', detail: '保存文章失败: ' + error.response?.data?.message, life: 2500})
  } finally {
    visible.value = false
    edittingIndex.value = -1
  }
}
</script> 