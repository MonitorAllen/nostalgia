<template>
    <div class="p-6">
      <h1 class="text-2xl font-bold mb-4">仪表盘</h1>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div class="bg-white p-6 rounded-lg shadow">
          <h2 class="text-lg font-semibold mb-2">系统状态</h2>
          <p class="text-green-600">运行正常</p>
        </div>
        <div class="bg-white p-6 rounded-lg shadow">
          <h2 class="text-lg font-semibold mb-2">最近活动</h2>
          <p class="text-gray-600">暂无活动</p>
        </div>
        <div class="bg-white p-6 rounded-lg shadow">
          <h2 class="text-lg font-semibold mb-2">系统信息</h2>
          <p class="text-gray-600">版本: 1.0.0</p>
        </div>
      </div>
    </div>
</template>

<script setup lang="ts">
</script> 