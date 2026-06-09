<script lang="ts" setup>
import { onBeforeUnmount, ref } from 'vue'
import { CheckCircle2, LoaderCircle, XCircle } from '@lucide/vue'
import { useUserStore } from '@/store/module/user'
import { useToast } from '@/composables/useToast'
import router from '@/router'

const { email_id, secret_code } = defineProps<{
  email_id: number
  secret_code: string
}>()

const isAuth = ref(false)
const isVerified = ref(false)
const timer = ref<number | null>(null)

if (email_id !== 0 && secret_code !== '') {
  const userStore = useUserStore()
  const toast = useToast()
  userStore
    .verifyEmail(email_id, secret_code)
    .then(() => {
      isAuth.value = true
      isVerified.value = true
      timer.value = window.setTimeout(() => {
        router.replace({ name: 'home' })
      }, 2000)
    })
    .catch((err) => {
      toast.add({
        severity: 'error',
        summary: '验证失败',
        detail: err.response?.data?.error || '邮箱验证失败',
        life: 3000,
      })
      isAuth.value = true
    })
}

onBeforeUnmount(() => {
  if (timer.value) clearTimeout(timer.value)
})
</script>

<template>
  <main class="grid min-h-screen place-items-center px-4">
    <section class="archive-surface w-full max-w-md rounded-[1.1rem] p-8 text-center">
      <LoaderCircle v-if="!isAuth" class="mx-auto h-10 w-10 animate-spin text-accent" />
      <CheckCircle2 v-else-if="isVerified" class="mx-auto h-10 w-10 text-accent" />
      <XCircle v-else class="mx-auto h-10 w-10 text-danger" />
      <h1 class="mt-4 text-2xl font-black">
        {{ !isAuth ? '账号验证中' : isVerified ? '账号验证成功' : '账号验证失败' }}
      </h1>
      <p class="m-0 mt-2 text-sm text-muted-foreground">
        {{ isVerified ? '即将回到首页。' : '请确认链接是否完整或重新发送验证邮件。' }}
      </p>
    </section>
  </main>
</template>
