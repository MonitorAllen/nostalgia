<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { LogIn } from '@lucide/vue'
import { useUserStore } from '@/store/module/user'
import { useToast } from '@/composables/useToast'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'

const user = ref({
  username: '',
  password: '',
})

const isLoginDisabled = computed(() => !user.value.username || !user.value.password)
const userStore = useUserStore()
const router = useRouter()
const toast = useToast()

const handleLogin = () => {
  userStore
    .login(user.value)
    .then(() => {
      toast.add({
        severity: 'success',
        summary: `欢迎回来，${userStore.userInfo.full_name}`,
        detail: '你已经成功登录。',
        life: 3000,
      })
      router.replace({ name: 'home' })
    })
    .catch((err: any) => {
      toast.add({
        severity: 'error',
        summary: '登录失败',
        detail: err.response?.data?.error || '请检查账号或密码',
        life: 3000,
      })
    })
}
</script>

<template>
  <main class="grid min-h-screen place-items-center px-4 py-10">
    <section class="archive-surface w-full max-w-md rounded-[1.1rem] p-6">
      <div class="mb-6 flex items-center gap-3">
        <span class="archive-glass grid h-11 w-11 place-items-center rounded-full">
          <LogIn class="h-5 w-5 text-accent" />
        </span>
        <div>
          <h1 class="m-0 text-2xl font-black">登录 Nostalgia</h1>
          <p class="m-0 text-sm text-muted-foreground">欢迎回来，继续阅读与评论。</p>
        </div>
      </div>

      <form class="space-y-4" @submit.prevent="handleLogin">
        <label class="block space-y-2">
          <span class="text-sm font-bold">用户名</span>
          <AppInput id="username" v-model="user.username" autocomplete="username" />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-bold">密码</span>
          <AppInput id="password" v-model="user.password" type="password" autocomplete="current-password" />
        </label>
        <AppButton class="w-full" type="submit" :disabled="isLoginDisabled">登录</AppButton>
      </form>

      <p class="m-0 mt-5 text-center text-sm text-muted-foreground">
        还没有账号？
        <RouterLink :to="{ name: 'register' }" class="font-bold text-accent">去注册</RouterLink>
      </p>
    </section>
  </main>
</template>
