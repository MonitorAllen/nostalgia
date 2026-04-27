<template>
  <div class="table-box">
    <ProTable
      ref="proTable"
      title="分类列表"
      :columns="columns"
      :request-api="getCategoryListApiWrapper"
      :pagination="false"
    >
      <template #tableHeader>
        <el-button type="primary" :icon="CirclePlus" @click="openDialog()">新增分类</el-button>
      </template>

      <template #operation="scope">
        <el-button type="primary" link :icon="Edit" @click="openDialog(scope.row)">编辑</el-button>
        <el-button type="danger" link :icon="Delete" @click="deleteCategory(scope.row)">
          删除
        </el-button>
      </template>
    </ProTable>

    <el-dialog
      v-model="dialogVisible"
      :title="currentCategory.id ? '编辑分类' : '新增分类'"
      width="400px"
    >
      <el-form :model="currentCategory" label-width="80px">
        <el-form-item label="分类名称" required>
          <el-input v-model="currentCategory.name" placeholder="请输入分类名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveCategory">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts" name="categoryManage">
import { ref } from 'vue'
import { ColumnProps } from '@/components/ProTable/interface'
import ProTable from '@/components/ProTable/index.vue'
import { CirclePlus, Delete, Edit } from '@element-plus/icons-vue'
import {
  getCategoryListApi,
  createCategoryApi,
  updateCategoryApi,
  deleteCategoryApi,
} from '@/api/modules/categories'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Nostalgia } from '@/api/interface/nostalgia'

const proTable = ref()

// Wrapper to match ProTable expectation
const getCategoryListApiWrapper = async () => {
  const res = await getCategoryListApi()
  return res
}

const columns: ColumnProps<Nostalgia.Category>[] = [
  { type: 'index', label: '#', width: 80 },
  { prop: 'name', label: '分类名称' },
  { prop: 'article_count', label: '文章数量' },
  { prop: 'created_at', label: '创建时间' },
  { prop: 'operation', label: '操作', fixed: 'right' },
]

const dialogVisible = ref(false)
const currentCategory = ref<Partial<Nostalgia.Category>>({})

const openDialog = (row?: Nostalgia.Category) => {
  if (row) {
    currentCategory.value = { ...row }
  } else {
    currentCategory.value = { name: '' }
  }
  dialogVisible.value = true
}

const saveCategory = async () => {
  if (!currentCategory.value.name) return ElMessage.warning('请输入分类名称')

  if (currentCategory.value.id) {
    await updateCategoryApi({ id: currentCategory.value.id, name: currentCategory.value.name })
    ElMessage.success('更新成功')
  } else {
    await createCategoryApi({ name: currentCategory.value.name })
    ElMessage.success('创建成功')
  }
  dialogVisible.value = false
  proTable.value?.getTableList()
}

const deleteCategory = async (row: Nostalgia.Category) => {
  await ElMessageBox.confirm('确认删除该分类吗？删除后文章分类将恢复默认。', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
  await deleteCategoryApi(row.id)
  ElMessage.success('删除成功')
  proTable.value?.getTableList()
}
</script>
