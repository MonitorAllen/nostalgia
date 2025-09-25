<template>
    <div class="flex flex-column w-full shadow-1 border-round px-3 pt-2">
      <div class="flex justify-content-between align-items-center mb-2">
        <h1 class="text-xl font-medium m-0">文章管理</h1>
        <router-link :to="{name: 'editor'}" target="_blank">
          <Button label="新建文章" icon="pi pi-plus" size="small" severity="primary"/>
        </router-link>
      </div>
      <div class="flex flex-column h-full">
        <DataTable
          :value="articles"
          size="small"
          showGridlines
          stripedRows
          responsiveLayout="scroll"
          class="p-datatable-sm mb-2">
          <Column field="title" header="标题" style="min-width: 140px;"></Column>
          <Column field="summary" header="摘要" style="min-width: 120px;">
            <template #body="slotProps">
              <div class="">{{ slotProps.data.summary.length > 50 ? slotProps.data.summary.substring(0, 30) + '...' : slotProps.data.summary }}</div>
            </template>
          </Column>
          <Column field="likes" header="点赞" style="width: 50px; min-width: 50px;"></Column>
          <Column field="views" header="浏览量" style="width: 65px; min-width: 65px;"></Column>
          <Column field="category_name" header="分类" style="width: 65px; min-width: 65px;"></Column>
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
              <Button label="删除" size="small" severity="danger" @click="deleteRecord(slotProps.data.id)"/>
            </template>
          </Column>
        </DataTable>

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
    </div>

    <Dialog v-model:visible="visible" @after-hide="onDialogHide" modal header="编辑文章信息" :style="{ width: '45rem' }">
      <Form :initialValues="formInitialValues" @submit="onUpdateArticle" class="flex flex-column">
        <InputText type="hidden" name="id"></InputText>
        <div class="flex items-center gap-4 mb-4">
            <label for="title" class="font-semibold w-24">标题</label>
            <InputText id="title" name="title" class="flex-auto"/>
        </div>
        <div class="flex flex-row gap-4 mb-4">
            <label for="summary" class="font-semibold w-24">摘要</label>
            <Textarea id="summary" name="summary" autoResize class="flex-auto"/>
        </div>
        <div class="flex flex-row justify-content-between gap-4 mb-4">
          <label for="category" class="font-semibold w-24">分类</label>
          <Select id="category" name="category_id" v-model="selectedCategory" :options="categories" optionLabel="name" optionValue="id"
          class="flex-auto"/>
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

    <ConfirmDialog></ConfirmDialog>
</template>

<script setup lang="ts">
import {ref, onMounted, computed} from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Paginator, {type PageState} from 'primevue/paginator'
import Dialog from 'primevue/dialog';
import InputText from 'primevue/inputtext';
import Textarea from 'primevue/textarea';
import ToggleSwitch from 'primevue/toggleswitch';
import Select from 'primevue/select';
import {Form, type FormSubmitEvent} from '@primevue/forms';
import { useConfirm } from "primevue/useconfirm";
import { type Article } from '@/stores/article'
import { useToast } from 'primevue/usetoast'
import ConfirmDialog from 'primevue/confirmdialog';
import format from '@/util/date'
import {
  deleteArticle,
  fetchArticleById, listAllArticles,
  updateArticle,
  type UpdateArticleRequest
} from '@/api/articles'
import {listAllCategories} from "@/api/category.ts";
import type {Category} from "@/types/category.ts";

const toast = useToast()

const articles = ref<Article[]>([])
const page = ref(1)
const limit = ref(10)
const count = ref(0)

const fetchAllArticles = async (page: number, limit: number) => {
  try {
    const resp = await listAllArticles({page, limit})
    articles.value = resp.data.articles
    count.value = parseInt(resp.data.count)
  } catch (error: any) {
    toast.add({ severity: 'error', summary: '错误', detail: '获取文章列表失败: ' + error.response.data.error, life: 2500})
  }
}

const first = ref(0)

onMounted( () => {
  fetchAllArticles(page.value, limit.value)
})

const onPage = (event: PageState) => {
  first.value = event.first
  page.value = event.page + 1
  limit.value = event.rows
  fetchAllArticles(page.value, limit.value)
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

// 计算表单初始值
const formInitialValues = computed(() => {
  if (!article.value) return {}

  return {
    id: article.value.id,
    title: article.value.title || '',
    summary: article.value.summary || '',
    is_publish: article.value.is_publish || false,
  }
})

const visible = ref(false)
const article = ref<Article | null>(null)
const categories = ref<Category[]>([])
const selectedCategory = ref('')
const editingIndex = ref(-1)

const editArticle = async (index: number, id: string) => {
  try {
    const [articleRes, categoryRes] = await Promise.all([fetchArticleById({id, needContent:false}), listAllCategories()]);
    categories.value = categoryRes.data.categories
    article.value = articleRes.data.article
    selectedCategory.value = articleRes.data.article.category_id
    visible.value = true
    editingIndex.value = index
  } catch (error: any) {
    toast.add({severity: 'error', summary: '错误', detail: '获取文章信息失败: ' + error.response?.data?.message, life: 2500})
  }
}

const onUpdateArticle = async (e: FormSubmitEvent<Record<string, any>>) => {
  const values = e.values
  values.category_id = parseInt(values.category_id)
  try {
    await updateArticle(values as UpdateArticleRequest)
    if (editingIndex.value != -1) {
        toast.add({severity: 'success', summary: '成功', detail: '保存文章成功', life: 2500})
    }
  } catch (error: any) {
    toast.add({severity: 'error', summary: '错误', detail: '保存文章失败: ' + error.response?.data?.message, life: 2500})
  } finally {
    visible.value = false
    editingIndex.value = -1
    article.value = null
    selectedCategory.value = ''
  }
}

const onDialogHide = () => {
  selectedCategory.value = ''
}

const confirm = useConfirm();

const deleteRecord = (id: string) => {
  confirm.require({
    message: '确定删除这篇文章吗？(数据无法恢复)',
    header: '警告',
    icon: 'pi pi-info-circle',
    rejectLabel: '取消',
    acceptProps: {
      label: '确定',
      severity: 'danger'
    },
    accept: async () => {
      try {
        await deleteArticle({id})
        toast.add({severity: 'success', summary: '成功', detail: '删除成功', life: 3000});
        await fetchAllArticles(page.value, limit.value)
      } catch (error: any) {
        toast.add({severity: 'error', summary: '失败', detail: '删除失败: ' + error.response.data.error, life: 3000});
      }
    }
  });
}
</script>
