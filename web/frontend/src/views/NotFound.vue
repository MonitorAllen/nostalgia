<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'

const router = useRouter()

const goBack = () => {
  router.go(-1)
}

const goHome = () => {
  router.push('/')
}
</script>

<template>
  <div class="flex flex-column align-items-center justify-content-center min-h-screen surface-ground px-4">
    <div class="mb-5 floating-icon">
      <span class="text-green-500 font-bold" style="font-size: 8rem; line-height: 1">4</span>
      <i class="pi pi-compass text-green-500 mx-2" style="font-size: 6rem; animation: spin 10s linear infinite;"></i>
      <span class="text-green-500 font-bold" style="font-size: 8rem; line-height: 1">4</span>
    </div>

    <div class="text-center mb-6">
      <h1 class="text-900 font-bold text-4xl mb-3">页面未找到</h1>
      <p class="text-600 text-xl line-height-3 m-0">
        抱歉，您访问的页面可能已被删除、更名<br>或者暂时不可用。
      </p>
    </div>

    <div class="flex gap-3">
      <Button
          label="返回上一页"
          icon="pi pi-arrow-left"
          outlined
          severity="secondary"
          @click="goBack"
      />
      <Button
          label="回到首页"
          icon="pi pi-home"
          severity="success"
          @click="goHome"
      />
    </div>
  </div>
</template>

<style scoped>
/* 简单的浮动动画 */
.floating-icon {
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0% { transform: translateY(0px); }
  50% { transform: translateY(-15px); }
  100% { transform: translateY(0px); }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 覆盖 PrimeVue 默认 success 按钮颜色以完全匹配你的 #10b981 */
:deep(.p-button.p-button-success) {
  background: #10b981;
  border-color: #10b981;
}

:deep(.p-button.p-button-success:hover) {
  background: #059669;
  border-color: #059669;
}
</style>