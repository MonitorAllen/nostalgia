<template>
    <div class="flex flex-column w-full shadow-1 border-round px-3 pt-2">
      <div class="flex justify-content-between align-items-center mb-2">
        <h1 class="text-xl font-medium m-0">分类管理</h1>
        <Button label="新增分类" icon="pi pi-plus" size="small" severity="primary" @click="createCategoryDialogVisible = true"/>
      </div>
      <div class="flex flex-column h-full">
        <DataTable
          :value="categories"
          size="small"
          showGridlines
          stripedRows
          responsiveLayout="scroll"
          class="p-datatable-sm mb-2">
          <Column field="name" header="名称" style="min-width: 140px;"></Column>
          <Column field="created_at" header="创建时间" style="width: 180px; min-width: 172px">
            <template #body="slotProps">
              {{ format(slotProps.data.created_at, 'YYYY-MM-DD HH:mm') }}
            </template>
          </Column>
          <Column field="updated_at" header="更新时间" style="width: 180px; min-width: 172px;">
            <template #body="slotProps">
              {{ format(slotProps.data.updated_at, 'YYYY-MM-DD HH:mm') }}
            </template>
          </Column>
          <Column header="操作" :exportable="false" style="width: 15rem; min-width:15rem">
            <template #body="slotProps">
              <Button label="编辑" size="small" severity="info" class="mr-2" @click="editCategory(slotProps.index)"/>
              <Button label="删除" size="small" severity="danger" @click="onDeleteCategory(slotProps.data.id)"/>
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

    <Dialog v-model:visible="editCategoryDialogVisible" modal header="编辑分类" :style="{ width: '20rem' }">
      <Form :initialValues="formInitialValues" @submit="onUpdateCategory" class="flex flex-column">
        <InputText type="hidden" name="id"></InputText>
        <div class="flex flex-row align-items-center gap-2 mb-4">
            <label for="name" class="font-semibold">名称</label>
            <InputText id="name" name="name" size="small" class="flex-auto"/>
        </div>
        <div class="flex justify-content-end gap-2">
            <Button type="button" label="取消" severity="secondary" @click="editCategoryDialogVisible = false"></Button>
            <Button type="submit" label="保存"></Button>
        </div>
      </Form>
    </Dialog>

  <Dialog v-model:visible="createCategoryDialogVisible" modal header="新增分类">
    <div class="flex flex-column">
      <div class="flex flex-row align-items-center gap-4 mb-4">
        <label for="name" class="font-semibold w-24">名称</label>
        <InputText id="name" v-model="createCategoryName" size="small" class="flex-auto"/>
      </div>
      <div class="flex justify-content-end gap-2">
        <Button type="button" label="取消" severity="secondary" @click="createCategoryDialogVisible = false"></Button>
        <Button type="submit" label="提交" @click="onCreateCategory"></Button>
      </div>
    </div>
  </Dialog>

    <ConfirmDialog></ConfirmDialog>
</template>

<script setup lang="ts">
import {ref, onMounted, computed} from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Paginator, {type PageState} from 'primevue/paginator'
import Dialog from 'primevue/dialog';
import InputText from 'primevue/inputtext';
import {Form, type FormSubmitEvent} from '@primevue/forms';
import { useConfirm } from "primevue/useconfirm";
import { useToast } from 'primevue/usetoast'
import ConfirmDialog from 'primevue/confirmdialog';
import format from '@/util/date'
import {
  createCategory,
  deleteCategory,
  listCategories,
  updateCategory,
  type UpdateCategoryRequest
} from "@/api/category.ts";
import type {Category} from "@/types/category.ts";

const toast = useToast()

const categories = ref<Category[]>([])
const page = ref(1)
const limit = ref(10)
const count = ref(0)
const first = ref(0)

const fetchCategories = async (page: number, limit: number) => {
  try {
    const resp = await listCategories({page, limit})
    categories.value = resp.data.categories
    count.value = parseInt(resp.data.count)
  } catch (error: any) {
    toast.add({ severity: 'error', summary: '错误', detail: '获取分类列表失败: ' + error.response.data.error, life: 2500})
  }
}

onMounted( () => {
  fetchCategories(page.value, limit.value)
})

const onPage = (event: PageState) => {
  first.value = event.first
  page.value = event.page + 1
  limit.value = event.rows
  fetchCategories(page.value, limit.value)
}

const editCategoryDialogVisible = ref(false)
const category = ref<Category | null>(null)
const editingIndex = ref(-1)

// 计算表单初始值
const formInitialValues = computed(() => {
  if (!category.value) return {}

  return {
    id: category.value.id,
    name: category.value.name || ''
}
})

const editCategory = async (index: number) => {
    category.value = categories.value[index]
    editCategoryDialogVisible.value = true
    editingIndex.value = index
}

const onUpdateCategory = async (e: FormSubmitEvent<Record<string, any>>) => {
  const values = e.values
  try {
    await updateCategory(values as UpdateCategoryRequest)
    if (editingIndex.value != -1) {
        toast.add({severity: 'success', summary: '成功', detail: '修改分类成功', life: 2500})
    }
    await fetchCategories(1, limit.value)
  } catch (error: any) {
    toast.add({severity: 'error', summary: '错误', detail: '修改分类失败: ' + error.response?.data?.message, life: 2500})
  } finally {
    editCategoryDialogVisible.value = false
    editingIndex.value = -1
    category.value = null
  }
}

const confirm = useConfirm();

const onDeleteCategory = (id: number) => {
  confirm.require({
    message: '确定删除这个分类吗？(文章分类恢复默认)',
    header: '警告',
    icon: 'pi pi-info-circle',
    rejectLabel: '取消',
    acceptProps: {
      label: '确定',
      severity: 'danger'
    },
    accept: async () => {
      try {
        await deleteCategory({id})
        toast.add({severity: 'success', summary: '成功', detail: '删除分类成功', life: 3000});
        await fetchCategories(page.value, limit.value)
      } catch (error: any) {
        toast.add({severity: 'error', summary: '失败', detail: '删除分类失败: ' + error.response.data.message, life: 3000});
      }
    }
  });
}

const createCategoryDialogVisible = ref(false)
const createCategoryName = ref('')
const onCreateCategory = async () => {
  if (createCategoryName.value === '') {
    toast.add({severity: 'warn', summary: '提示', detail: '名称不能为空', life: 3000});
    return
  }
  try {
    await createCategory({name: createCategoryName.value})
    await fetchCategories(1, limit.value)
    toast.add({severity: 'success', summary: '成功', detail: '新增分类成功', life: 3000});
  } catch (error: any) {
    toast.add({severity: 'error', summary: '失败', detail: '新增分类失败: ' + error.response.data.message, life: 3000});
  } finally {
    createCategoryDialogVisible.value = false
    createCategoryName.value = ''
  }
}
</script>
