<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { UserPlus } from '@lucide/vue'
import { useUserStore } from '@/store/module/user'
import { useToast } from '@/composables/useToast'
import type { RegisterRequest } from '@/types/request/user'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'

const user = ref<RegisterRequest>({
  username: '',
  email: '',
  full_name: '',
  password: '',
})

const isRegisterDisabled = computed(
  () => !user.value.email || !user.value.username || !user.value.full_name || !user.value.password,
)
const userStore = useUserStore()
const router = useRouter()
const toast = useToast()

const handleRegister = () => {
  userStore
    .register(user.value)
    .then(() => {
      toast.add({ severity: 'success', summary: '注册成功', detail: '欢迎加入 Nostalgia。', life: 3000 })
      router.replace({ name: 'home' })
    })
    .catch((err: any) => {
      const detail =
        err.response?.status === 403
          ? '邮箱或用户名已存在'
          : err.response?.status === 400
            ? '注册参数有误'
            : '请稍后再试'
      toast.add({ severity: 'error', summary: '注册失败', detail, life: 3000 })
    })
}
</script>

<template>
  <main class="grid min-h-screen place-items-center px-4 py-10">
    <section class="archive-surface w-full max-w-md rounded-[1.1rem] p-6">
      <div class="mb-6 flex items-center gap-3">
        <span class="archive-glass grid h-11 w-11 place-items-center rounded-full">
          <UserPlus class="h-5 w-5 text-accent" />
        </span>
        <div>
          <h1 class="m-0 text-2xl font-black">创建 Nostalgia 账号</h1>
          <p class="m-0 text-sm text-muted-foreground">注册后可以评论、验证邮箱并继续阅读。</p>
        </div>
      </div>

      <form class="space-y-4" @submit.prevent="handleRegister">
        <label class="block space-y-2">
          <span class="text-sm font-bold">Email</span>
          <AppInput id="email" v-model="user.email" type="email" autocomplete="email" />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-bold">Username</span>
          <AppInput id="username" v-model="user.username" autocomplete="username" />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-bold">Full name</span>
          <AppInput id="full_name" v-model="user.full_name" autocomplete="name" />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-bold">Password</span>
          <AppInput id="password" v-model="user.password" type="password" autocomplete="new-password" />
        </label>
        <AppButton class="w-full" type="submit" :disabled="isRegisterDisabled">注册</AppButton>
      </form>

      <p class="m-0 mt-5 text-center text-sm text-muted-foreground">
        已经有账号？
        <RouterLink :to="{ name: 'login' }" class="font-bold text-accent">去登录</RouterLink>
      </p>
    </section>
  </main>
</template>
