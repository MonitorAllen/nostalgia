<template>
  <div class="login-container">
    <!-- 左侧猫咪动画 -->
    <div class="login-left">
      <div class="w-full flex align-items-center justify-content-center" style="max-width: 32rem">
        <svg viewBox="0 0 400 400" class="w-full h-full">
          <!-- 猫头 -->
          <circle cx="200" cy="200" r="120" fill="#F8F9FA" />
          <!-- 猫耳朵 -->
          <path d="M130 140 C70 60 110 40 170 120" fill="#F8F9FA" />
          <path d="M270 140 C330 60 290 40 230 120" fill="#F8F9FA" />
          <!-- 猫耳朵内部 -->
          <path d="M120 105 Q95 65 140 95" fill="none" stroke="#FFB6C1" stroke-width="2.5" />
          <path d="M280 105 Q305 65 260 95" fill="none" stroke="#FFB6C1" stroke-width="2.5" />
          <!-- 猫脸 -->
          <!-- 左眼 -->
          <circle cx="160" cy="190" r="25" fill="#FFE4B5" />
          <circle cx="160" cy="190" r="12" fill="#000" />
          <!-- 右眼 -->
          <defs>
            <radialGradient id="eyeGradient" cx="50%" cy="50%" r="50%" fx="50%" fy="50%">
              <stop offset="0%" style="stop-color:#001F7A" />
              <stop offset="15%" style="stop-color:#0033CC" />
              <stop offset="25%" style="stop-color:#0047E1" />
              <stop offset="35%" style="stop-color:#0052FF" />
              <stop offset="45%" style="stop-color:#1A75FF" />
              <stop offset="55%" style="stop-color:#3399FF" />
              <stop offset="65%" style="stop-color:#66B2FF" />
              <stop offset="72%" style="stop-color:#80BFFF" />
              <stop offset="80%" style="stop-color:#99CCFF" />
              <stop offset="90%" style="stop-color:#BFDFFF" />
              <stop offset="100%" style="stop-color:#E6F3FF" />
            </radialGradient>
          </defs>
          <circle cx="240" cy="190" r="25" fill="url(#eyeGradient)" />
          <circle cx="240" cy="190" r="12" fill="#000" />
          <!-- 猫鼻子 -->
          <path d="M190 230 L210 230 L200 242 Z" fill="#FFB6C1" />
          <!-- 猫嘴 -->
          <path d="M180 255 Q200 275 220 255" fill="none" stroke="#FFB6C1" stroke-width="4" />
          <!-- 猫胡子 -->
          <line x1="150" y1="245" x2="80" y2="225" stroke="#FFFFFF" stroke-width="3" />
          <line x1="150" y1="255" x2="80" y2="255" stroke="#FFFFFF" stroke-width="3" />
          <line x1="150" y1="265" x2="80" y2="285" stroke="#FFFFFF" stroke-width="3" />
          <line x1="250" y1="245" x2="320" y2="225" stroke="#FFFFFF" stroke-width="3" />
          <line x1="250" y1="255" x2="320" y2="255" stroke="#FFFFFF" stroke-width="3" />
          <line x1="250" y1="265" x2="320" y2="285" stroke="#FFFFFF" stroke-width="3" />
        </svg>
      </div>
    </div>

    <!-- 右侧登录表单 -->
    <div class="login-right">
      <div class="surface-card shadow-2 border-round w-full" style="max-width: 32rem">
        <div class="text-center p-4 sm:p-5 md:p-6">
          <div class="text-900 text-3xl font-medium mb-2">欢迎回来</div>
          <span class="text-600 font-medium">请登录您的账号</span>
        </div>

        <form @submit.prevent="handleLogin" class="p-4 sm:p-5 md:p-6 flex flex-column gap-4">
          <div class="flex flex-column gap-2">
            <label for="username" class="text-900 font-medium">用户名</label>
            <InputText
              id="username"
              v-model="form.username"
              type="text"
              class="w-full"
              :class="{ 'p-invalid': submitted && !form.username }"
              placeholder="请输入用户名"
            />
            <small v-if="submitted && !form.username" class="p-error">用户名不能为空</small>
          </div>

          <div class="flex flex-column gap-2">
            <label for="password" class="text-900 font-medium">密码</label>
            <Password
              id="password"
              v-model="form.password"
              class="w-full"
              :class="{ 'p-invalid': submitted && !form.password }"
              :feedback="false"
              toggleMask
              placeholder="请输入密码"
            />
            <small v-if="submitted && !form.password" class="p-error">密码不能为空</small>
          </div>

          <div class="flex align-items-center justify-content-between mb-4">
            <div class="flex align-items-center">
              <Checkbox
                v-model="rememberMe"
                binary
                class="mr-2"
              />
              <label for="rememberMe" class="text-900">记住我</label>
            </div>
            <a class="font-medium text-primary no-underline hover:underline cursor-pointer">忘记密码？</a>
          </div>

          <Button
            type="submit"
            label="登录"
            :loading="loading"
            class="w-full"
          />
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Checkbox from 'primevue/checkbox'
import { useToast } from 'primevue/usetoast'

const router = useRouter()
const authStore = useAuthStore()
const toast = useToast()

if (authStore.isAuthenticated) {
  router.replace('/')
}

const form = ref({
  username: '',
  password: ''
})

const loading = ref(false)
const submitted = ref(false)
const rememberMe = ref(false)

const handleLogin = async () => {
  submitted.value = true

  if (!form.value.username || !form.value.password) {
    return
  }

  try {
    loading.value = true
    await authStore.login(form.value)
    toast.add({
      severity: 'success',
      summary: '登录成功',
      detail: '欢迎回来！',
      life: 3000
    })
    router.push('/')
  } catch (error: any) {
    console.error('登录失败:', error)
    toast.add({
      severity: 'error',
      summary: '登录失败',
      detail: error.response?.data.message || '用户名或密码错误',
      life: 3000
    })
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
:deep(.p-password input) {
  width: 100%;
}

.login-container {
  position: fixed;
  inset: 0;
  display: grid;
  grid-template-columns: 1fr;
  min-height: 100vh;
  background: #F5F2EA;
}

.login-left {
  display: none;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-400) 100%);
  padding: 1.5rem;
  place-items: center;
}

.login-right {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  min-width: 300px;
  overflow-x: auto;
}

@media screen and (min-width: 992px) {
  .login-container {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  
  .login-left {
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style> 