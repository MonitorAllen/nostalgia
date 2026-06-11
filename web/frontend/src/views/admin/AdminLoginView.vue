<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ArchivePanel from '@/components/ui/ArchivePanel.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { ADMIN_ARTICLES_PATH } from '@/admin/adminRoutes'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/store/module/auth'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const isSubmitting = ref(false)
const isCheckingSession = ref(true)
const errorMessage = ref('')

const redirectPath = computed(() => {
  const redirect = route.query.redirect

  if (typeof redirect !== 'string') {
    return ADMIN_ARTICLES_PATH
  }

  const resolved = router.resolve(redirect)
  const isAdminProtectedRoute =
    resolved.name !== 'adminLogin' && resolved.matched.some((record) => record.meta.requiresAdmin)

  return isAdminProtectedRoute ? resolved.fullPath : ADMIN_ARTICLES_PATH
})

const isSubmitDisabled = computed(() => {
  return isCheckingSession.value || isSubmitting.value || !username.value || !password.value
})

const extractErrorDetail = (error: unknown) => {
  if (error instanceof Error && error.message === 'Admin role required') {
    return '当前账号没有后台权限'
  }

  if (typeof error === 'object' && error && 'response' in error) {
    const response = (
      error as { response?: { data?: { error?: string; message?: string } | string } }
    ).response
    const data = response?.data

    if (typeof data === 'string' && data) return data
    if (data?.error) return data.error
    if (data?.message) return data.message
  }

  return '管理员账号或密码不正确'
}

const redirectToAdmin = async () => {
  await router.replace(redirectPath.value)
}

const handleSubmit = async () => {
  if (isSubmitDisabled.value) return

  errorMessage.value = ''
  isSubmitting.value = true

  try {
    const response = await authStore.login({
      username: username.value,
      password: password.value
    })

    if (response.data.user.role !== 'admin') {
      authStore.clearTokens()
      throw new Error('Admin role required')
    }

    await redirectToAdmin()
  } catch (error) {
    errorMessage.value = extractErrorDetail(error)
    toast.add({
      severity: 'error',
      summary: '登录失败',
      detail: errorMessage.value,
      life: 3000
    })
  } finally {
    isSubmitting.value = false
  }
}

onMounted(async () => {
  const authenticated = await authStore.ensureAdminAuthenticated()
  isCheckingSession.value = false

  if (authenticated) {
    await redirectToAdmin()
  }
})
</script>

<template>
  <section class="grid min-h-dvh place-items-center px-4 py-10" aria-label="管理员登录">
    <ArchivePanel class="w-full max-w-md" :glass="false">
      <div class="mb-6 flex items-center gap-3">
        <span class="archive-glass grid size-12 place-items-center rounded-full">
          <img src="/logo.svg" alt="Nostalgia" class="size-8" />
        </span>
        <div>
          <h1 class="m-0 text-2xl font-black text-balance">管理入口</h1>
          <p class="m-0 text-sm text-muted-foreground text-pretty">进入写作与内容维护工作台。</p>
        </div>
      </div>

      <form
        class="space-y-4"
        :aria-busy="isSubmitting || isCheckingSession"
        @submit.prevent="handleSubmit"
      >
        <label class="block space-y-2">
          <span class="text-sm font-bold">用户名</span>
          <AppInput
            id="admin-username"
            v-model="username"
            autocomplete="username"
            :disabled="isSubmitting || isCheckingSession"
            :aria-invalid="Boolean(errorMessage)"
            :aria-describedby="errorMessage ? 'admin-login-error' : undefined"
          />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-bold">密码</span>
          <AppInput
            id="admin-password"
            v-model="password"
            type="password"
            autocomplete="current-password"
            :disabled="isSubmitting || isCheckingSession"
            :aria-invalid="Boolean(errorMessage)"
            :aria-describedby="errorMessage ? 'admin-login-error' : undefined"
          />
        </label>

        <p
          v-if="errorMessage"
          id="admin-login-error"
          class="m-0 rounded-archive border border-danger/30 bg-danger/10 px-3 py-2 text-sm font-semibold text-danger"
          aria-live="polite"
        >
          {{ errorMessage }}
        </p>

        <AppButton class="w-full" type="submit" :disabled="isSubmitDisabled">
          {{ isCheckingSession ? '确认会话' : isSubmitting ? '登录中' : '登录' }}
        </AppButton>
      </form>
    </ArchivePanel>
  </section>
</template>
