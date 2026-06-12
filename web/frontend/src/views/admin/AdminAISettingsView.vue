<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Bot, CheckCircle2, KeyRound, RefreshCw, Save, ShieldCheck, XCircle } from '@lucide/vue'
import { getAdminAIConfig, updateAdminAIConfig } from '@/admin/api/adminAiApi'
import type { AdminAIConfigResponse } from '@/admin/types'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const toast = useToast()
const config = ref<AdminAIConfigResponse | null>(null)
const loading = ref(false)
const saving = ref(false)
const apiKeyDraft = ref('')
const clearApiKey = ref(false)
const form = ref({
  provider: 'openai_compatible',
  baseUrl: '',
  model: '',
  timeout: '30s',
  maxInputChars: '6000',
  maxContextChars: '4000',
  maxSuggestions: '3',
  enabled: true
})

const providerLabel = computed(() => config.value?.provider || '未设置')
const modelLabel = computed(() => config.value?.model || '未设置')
const baseUrlLabel = computed(() => config.value?.base_url || '未设置')
const sourceLabel = computed(() =>
  config.value?.source === 'runtime_env'
    ? '运行时环境变量'
    : config.value?.source === 'database'
      ? '后台实时配置'
      : config.value?.source || '未设置'
)
const canSubmit = computed(() => !loading.value && !saving.value)

const syncForm = (nextConfig: AdminAIConfigResponse | null) => {
  form.value = {
    provider: nextConfig?.provider || 'openai_compatible',
    baseUrl: nextConfig?.base_url || '',
    model: nextConfig?.model || '',
    timeout: nextConfig?.timeout || '30s',
    maxInputChars: String(nextConfig?.max_input_chars || 6000),
    maxContextChars: String(nextConfig?.max_context_chars || 4000),
    maxSuggestions: String(nextConfig?.max_suggestions || 3),
    enabled: nextConfig?.enabled ?? true
  }
  apiKeyDraft.value = ''
  clearApiKey.value = false
}

const toPositiveNumber = (value: string, fallback: number) => {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback
}

const toNonNegativeNumber = (value: string, fallback: number) => {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : fallback
}

const loadConfig = async () => {
  if (loading.value) return
  loading.value = true

  try {
    const response = await getAdminAIConfig()
    config.value = response.data
    syncForm(response.data)
  } catch {
    config.value = null
    syncForm(null)
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  if (!canSubmit.value) return
  saving.value = true

  try {
    const response = await updateAdminAIConfig({
      provider: form.value.provider.trim(),
      base_url: form.value.baseUrl.trim(),
      model: form.value.model.trim(),
      api_key: clearApiKey.value ? '' : apiKeyDraft.value.trim(),
      timeout: form.value.timeout.trim() || '30s',
      max_input_chars: toPositiveNumber(form.value.maxInputChars, 6000),
      max_context_chars: toNonNegativeNumber(form.value.maxContextChars, 4000),
      max_suggestions: toPositiveNumber(form.value.maxSuggestions, 3),
      enabled: form.value.enabled && !clearApiKey.value,
      clear_api_key: clearApiKey.value
    })
    config.value = response.data
    syncForm(response.data)
    toast.add({
      severity: 'success',
      summary: 'AI 配置已保存',
      detail: '新的提供商配置会在下一次润色请求中生效。',
      life: 2400
    })
  } finally {
    saving.value = false
  }
}

const openPolishTester = () => {
  void router.push({ name: 'adminArticleNew' })
}

onMounted(() => {
  void loadConfig()
})
</script>

<template>
  <main class="space-y-5">
    <header class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="m-0 text-2xl font-black leading-tight text-foreground">AI 提供商</h1>
          <AppBadge v-if="config" :tone="config.enabled ? 'accent' : 'warning'">
            {{ config.enabled ? '可用' : '未启用' }}
          </AppBadge>
        </div>
      </div>

      <div class="flex flex-wrap gap-2">
        <AppButton variant="secondary" :disabled="loading" @click="loadConfig">
          <RefreshCw class="size-4" aria-hidden="true" />
          {{ loading ? '刷新中' : '刷新状态' }}
        </AppButton>
        <AppButton @click="openPolishTester">
          <Bot class="size-4" aria-hidden="true" />
          测试润色
        </AppButton>
      </div>
    </header>

    <section v-if="loading && !config" class="archive-surface rounded-archive p-8 text-center">
      <p class="m-0 text-sm font-semibold text-muted-foreground">正在读取 AI 配置</p>
    </section>

    <section v-else class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]">
      <form class="archive-surface rounded-archive p-5" @submit.prevent="saveConfig">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h2 class="m-0 text-lg font-black text-foreground">实时配置</h2>
            <p class="m-0 mt-1 text-sm text-muted-foreground">
              保存后立即影响后台润色请求，API Key 不会回传到浏览器。
            </p>
          </div>
          <AppButton type="submit" :disabled="!canSubmit">
            <Save class="size-4" aria-hidden="true" />
            {{ saving ? '保存中' : '保存配置' }}
          </AppButton>
        </div>

        <div class="grid gap-3 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">Provider</span>
            <select
              v-model="form.provider"
              class="h-11 w-full rounded-full border border-border bg-surface px-4 text-sm font-semibold text-foreground transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/18 disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="!canSubmit"
            >
              <option value="openai_compatible">OpenAI Compatible</option>
            </select>
          </label>

          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">Model</span>
            <AppInput v-model="form.model" placeholder="gpt-4.1-mini" :disabled="!canSubmit" />
          </label>

          <label class="block space-y-2 md:col-span-2">
            <span class="text-sm font-bold text-foreground">Base URL</span>
            <AppInput
              v-model="form.baseUrl"
              placeholder="https://api.example.com/v1"
              :disabled="!canSubmit"
            />
          </label>

          <label class="block space-y-2 md:col-span-2">
            <span class="text-sm font-bold text-foreground">API Key</span>
            <AppInput
              v-model="apiKeyDraft"
              type="password"
              autocomplete="new-password"
              :placeholder="config?.api_key_configured ? '留空则保留当前密钥' : '输入新的 API Key'"
              :disabled="!canSubmit || clearApiKey"
            />
          </label>
        </div>

        <div class="mt-4 grid gap-3 sm:grid-cols-3">
          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">请求超时</span>
            <AppInput v-model="form.timeout" placeholder="30s" :disabled="!canSubmit" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">输入上限</span>
            <AppInput v-model="form.maxInputChars" type="number" :disabled="!canSubmit" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">候选数量</span>
            <AppInput v-model="form.maxSuggestions" type="number" :disabled="!canSubmit" />
          </label>
        </div>

        <div class="mt-4 grid gap-3 sm:grid-cols-2">
          <label class="flex items-start gap-3 rounded-archive border border-border bg-surface-raised p-3">
            <input
              v-model="form.enabled"
              type="checkbox"
              class="mt-1 size-4 rounded border-border accent-accent"
              :disabled="!canSubmit"
            />
            <span>
              <span class="block text-sm font-bold text-foreground">启用 AI 润色</span>
              <span class="block text-sm text-muted-foreground">关闭后保留配置，但润色请求会返回未配置。</span>
            </span>
          </label>

          <label class="flex items-start gap-3 rounded-archive border border-border bg-surface-raised p-3">
            <input
              v-model="clearApiKey"
              type="checkbox"
              class="mt-1 size-4 rounded border-border accent-danger"
              :disabled="!canSubmit"
            />
            <span>
              <span class="block text-sm font-bold text-foreground">清空 API Key</span>
              <span class="block text-sm text-muted-foreground">保存后移除当前密钥，并禁用可用状态。</span>
            </span>
          </label>
        </div>

        <dl class="m-0 mt-5 grid gap-3 sm:grid-cols-3">
          <div class="rounded-archive border border-border p-4">
            <dt class="text-xs font-bold text-muted-foreground">当前 Provider</dt>
            <dd class="m-0 mt-2 break-all text-sm font-black text-foreground">{{ providerLabel }}</dd>
          </div>
          <div class="rounded-archive border border-border p-4">
            <dt class="text-xs font-bold text-muted-foreground">当前 Model</dt>
            <dd class="m-0 mt-2 break-all text-sm font-black text-foreground">{{ modelLabel }}</dd>
          </div>
          <div class="rounded-archive border border-border p-4">
            <dt class="text-xs font-bold text-muted-foreground">当前 Base URL</dt>
            <dd class="m-0 mt-2 break-all text-sm font-black text-foreground">{{ baseUrlLabel }}</dd>
          </div>
        </dl>
      </form>

      <aside class="archive-surface h-max rounded-archive p-5">
        <h2 class="m-0 text-base font-black text-foreground">安全状态</h2>
        <div class="mt-4 space-y-3">
          <div class="flex items-start gap-3 rounded-archive border border-border bg-surface-raised p-3">
            <KeyRound class="mt-0.5 size-4 shrink-0 text-accent" aria-hidden="true" />
            <div class="min-w-0">
              <p class="m-0 text-sm font-bold text-foreground">API Key</p>
              <p class="m-0 mt-1 text-sm text-muted-foreground">
                {{ config?.api_key_configured ? '已配置，浏览器不可见' : '未配置' }}
              </p>
            </div>
            <CheckCircle2
              v-if="config?.api_key_configured"
              class="ml-auto size-4 shrink-0 text-accent"
              aria-hidden="true"
            />
            <XCircle v-else class="ml-auto size-4 shrink-0 text-warning" aria-hidden="true" />
          </div>

          <div class="flex items-start gap-3 rounded-archive border border-border bg-surface-raised p-3">
            <ShieldCheck class="mt-0.5 size-4 shrink-0 text-accent" aria-hidden="true" />
            <div>
              <p class="m-0 text-sm font-bold text-foreground">配置来源</p>
              <p class="m-0 mt-1 text-sm text-muted-foreground">{{ sourceLabel }}</p>
            </div>
          </div>
        </div>
      </aside>
    </section>
  </main>
</template>
