<template>
    <div class="p-6">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold">文章管理</h1>
        <button class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700">
          新建文章
        </button>
      </div>
      <div class="bg-white rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">标题</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">分类</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">发布时间</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">暂无数据</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">-</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">-</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">-</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">-</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.articles {
  padding: 1rem;
}
</style>