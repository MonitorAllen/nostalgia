<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'

const router = useRouter()

const goHome = () => {
  router.push('/')
}

const goLogin = () => {
  router.push('/login')
}
</script>

<template>
  <div class="flex flex-column align-items-center justify-content-center min-h-screen surface-ground px-4">
    <div class="flex align-items-center justify-content-center bg-green-50 border-circle mb-5"
         style="width: 120px; height: 120px; box-shadow: 0 0 0 15px rgba(16, 185, 129, 0.1)">
      <i class="pi pi-lock text-green-500 text-6xl"></i>
    </div>

    <div class="text-center mb-6">
      <h1 class="text-900 font-bold text-4xl mb-3">访问受限 (403)</h1>
      <p class="text-600 text-xl line-height-3 m-0 max-w-30rem">
        抱歉，您没有权限访问该页面。<br>如果您认为这是一个错误，请尝试重新登录或联系管理员。
      </p>
    </div>

    <div class="flex gap-3">
      <Button
          label="回到首页"
          icon="pi pi-home"
          text
          severity="secondary"
          @click="goHome"
      />
      <Button
          label="去登录"
          icon="pi pi-user"
          severity="success"
          @click="goLogin"
      />
    </div>
  </div>
</template>

<style scoped>
/* 同样匹配主题色 */
:deep(.p-button.p-button-success) {
  background: #10b981;
  border-color: #10b981;
}
:deep(.p-button.p-button-success:hover) {
  background: #059669;
  border-color: #059669;
}
</style>