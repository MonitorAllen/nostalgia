<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { KeyRound, RefreshCw, ShieldCheck } from '@lucide/vue'
import ArchivePanel from '@/components/ui/ArchivePanel.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { useToast } from '@/composables/useToast'
import {
  createSetupAdmin,
  getSetupStatus,
  type CreateSetupAdminPayload,
  type SetupStatus,
} from '@/service/setupService'

const router = useRouter()
const toast = useToast()

const status = ref<SetupStatus | null>(null)
const isLoading = ref(true)
const isSubmitting = ref(false)
const errorMessage = ref('')

const form = ref<CreateSetupAdminPayload>({
  setup_token: '',
  username: '',
  password: '',
  full_name: '',
  email: '',
})

const isSetupAvailable = computed(() => Boolean(status.value?.setup_available && !status.value.initialized))

const isSubmitDisabled = computed(() => {
  return (
    isLoading.value ||
    isSubmitting.value ||
    !isSetupAvailable.value ||
    !form.value.setup_token ||
    !form.value.username ||
    !form.value.password ||
    !form.value.full_name ||
    !form.value.email
  )
})

const extractErrorDetail = (error: unknown) => {
  if (typeof error === 'object' && error && 'response' in error) {
    const response = (error as { response?: { status?: number; data?: { error?: string; message?: string } | string } }).response
    const data = response?.data

    if (response?.status === 409) return '站点已经完成初始化'
    if (typeof data === 'string' && data) return data
    if (data?.error) return data.error
    if (data?.message) return data.message
  }

  return '初始化失败，请稍后再试'
}

const loadStatus = async () => {
  isLoading.value = true
  errorMessage.value = ''

  try {
    const response = await getSetupStatus()
    status.value = response.data
  } catch (error) {
    errorMessage.value = extractErrorDetail(error)
  } finally {
    isLoading.value = false
  }
}

const handleSubmit = async () => {
  if (isSubmitDisabled.value) return

  isSubmitting.value = true
  errorMessage.value = ''

  try {
    await createSetupAdmin(form.value)
    toast.add({
      severity: 'success',
      summary: '初始化完成',
      detail: '管理员账号已创建。',
      life: 3000,
    })
    await router.replace({ name: 'adminLogin' })
  } catch (error) {
    errorMessage.value = extractErrorDetail(error)

    if (typeof error === 'object' && error && 'response' in error) {
      const response = (error as { response?: { status?: number } }).response
      if (response?.status === 409) {
        await loadStatus()
      }
    }

    toast.add({
      severity: 'error',
      summary: '初始化失败',
      detail: errorMessage.value,
      life: 3000,
    })
  } finally {
    isSubmitting.value = false
  }
}

onMounted(loadStatus)
</script>

<template>
  <section class="grid min-h-dvh place-items-center px-4 py-10" aria-label="站点初始化">
    <ArchivePanel class="w-full max-w-lg" :glass="false">
      <div class="mb-6 flex items-center gap-3">
        <span class="archive-glass grid size-12 place-items-center rounded-full">
          <ShieldCheck class="size-5 text-accent" aria-hidden="true" />
        </span>
        <div class="min-w-0">
          <h1 class="m-0 text-2xl font-black text-balance">站点初始化</h1>
          <p class="m-0 text-sm text-muted-foreground text-pretty">创建第一个管理员账号。</p>
        </div>
      </div>

      <div v-if="isLoading" class="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
        <RefreshCw class="size-4 animate-spin" aria-hidden="true" />
        <span>正在确认状态</span>
      </div>

      <div v-else-if="status?.initialized" class="space-y-4">
        <p class="m-0 rounded-archive border border-border bg-muted px-3 py-2 text-sm font-semibold">
          站点已经完成初始化。
        </p>
        <RouterLink :to="{ name: 'adminLogin' }" class="inline-flex">
          <AppButton>进入管理入口</AppButton>
        </RouterLink>
      </div>

      <div v-else-if="!isSetupAvailable" class="space-y-4">
        <p class="m-0 rounded-archive border border-danger/30 bg-danger/10 px-3 py-2 text-sm font-semibold text-danger">
          {{ errorMessage || '初始化暂不可用' }}
        </p>
        <AppButton variant="secondary" @click="loadStatus">
          <RefreshCw class="size-4" aria-hidden="true" />
          <span>刷新状态</span>
        </AppButton>
      </div>

      <form v-else class="space-y-4" :aria-busy="isSubmitting" @submit.prevent="handleSubmit">
        <label class="block space-y-2">
          <span class="text-sm font-bold">Setup Token</span>
          <AppInput
            id="setup-token"
            v-model="form.setup_token"
            type="password"
            autocomplete="one-time-code"
            :disabled="isSubmitting"
          />
        </label>

        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-bold">用户名</span>
            <AppInput id="setup-username" v-model="form.username" autocomplete="username" :disabled="isSubmitting" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-bold">Email</span>
            <AppInput id="setup-email" v-model="form.email" type="email" autocomplete="email" :disabled="isSubmitting" />
          </label>
        </div>

        <label class="block space-y-2">
          <span class="text-sm font-bold">Full name</span>
          <AppInput id="setup-full-name" v-model="form.full_name" autocomplete="name" :disabled="isSubmitting" />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-bold">密码</span>
          <AppInput
            id="setup-password"
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            :disabled="isSubmitting"
          />
        </label>

        <p
          v-if="errorMessage"
          class="m-0 rounded-archive border border-danger/30 bg-danger/10 px-3 py-2 text-sm font-semibold text-danger"
          aria-live="polite"
        >
          {{ errorMessage }}
        </p>

        <AppButton class="w-full" type="submit" :disabled="isSubmitDisabled">
          <KeyRound class="size-4" aria-hidden="true" />
          <span>{{ isSubmitting ? '创建中' : '创建管理员' }}</span>
        </AppButton>
      </form>
    </ArchivePanel>
  </section>
</template>
