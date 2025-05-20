<template>
    <div class="surface-card h-full shadow-2 border-round p-3">
      <div class="flex justify-content-between align-items-center mb-4">
        <h1 class="text-xl font-medium m-0">用户管理</h1>
        <Button label="新建用户" icon="pi pi-plus" severity="primary" />
      </div>
      <DataTable
        :value="users"
        showGridlines
        stripedRows
        :paginator="true"
        :rows="10"
        paginatorTemplate="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
        :rowsPerPageOptions="[10,20,50]"
        currentPageReportTemplate="显示第 {first} 到 {last} 条记录，共 {totalRecords} 条"
        responsiveLayout="scroll"
        class="p-datatable-sm"
      >
        <Column field="username" header="用户名"></Column>
        <Column field="email" header="邮箱"></Column>
        <Column field="role" header="角色"></Column>
        <Column field="status" header="状态">
          <template #body="slotProps">
            <Tag :value="slotProps.data.status" :severity="getStatusSeverity(slotProps.data.status)" />
          </template>
        </Column>
        <Column field="createdAt" header="创建时间"></Column>
        <Column header="操作" :exportable="false" style="min-width:8rem">
          <template #body="slotProps">
            <Button icon="pi pi-pencil" outlined rounded severity="info" class="mr-2" />
            <Button icon="pi pi-trash" outlined rounded severity="danger" />
          </template>
        </Column>
      </DataTable>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Tag from 'primevue/tag'

interface User {
  username: string
  email: string
  role: string
  status: string
  createdAt: string
}

const users = ref<User[]>([
  {
    username: '暂无数据',
    email: '-',
    role: '-',
    status: '未知',
    createdAt: '-'
  }
])

const getStatusSeverity = (status: string) => {
  switch (status) {
    case '正常':
      return 'success'
    case '禁用':
      return 'danger'
    default:
      return 'info'
  }
}
</script> 