<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Bot, CheckCircle2, KeyRound, RefreshCw, Save, ShieldCheck, XCircle } from '@lucide/vue'
import { getAdminAIConfig, listAdminAIModels, updateAdminAIConfig } from '@/admin/api/adminAiApi'
import { getAIPolishModeLabel } from '@/admin/ai/polish'
import type { AdminAIConfigResponse, AdminAIPolishMode, AdminAIProtocol } from '@/admin/types'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const toast = useToast()
const config = ref<AdminAIConfigResponse | null>(null)
const loading = ref(false)
const saving = ref(false)
const modelsLoading = ref(false)
const modelOptions = ref<string[]>([])
const apiKeyDraft = ref('')
const clearApiKey = ref(false)
const promptModes: AdminAIPolishMode[] = [
  'improve',
  'shorten',
  'expand',
  'title_candidates',
  'summary_candidates'
]
const createPromptTemplateState = (): Record<AdminAIPolishMode, string> => ({
  improve: '',
  shorten: '',
  expand: '',
  title_candidates: '',
  summary_candidates: ''
})
const promptTemplates = ref<Record<AdminAIPolishMode, string>>(createPromptTemplateState())
const defaultPromptTemplates = ref<Record<AdminAIPolishMode, string>>(createPromptTemplateState())
const form = ref({
  provider: 'OpenAI Compatible',
  apiProtocol: 'chat/completions' as AdminAIProtocol,
  baseUrl: '',
  model: '',
  timeout: '30s',
  maxInputChars: '6000',
  maxContextChars: '4000',
  maxSuggestions: '3',
  enabled: true
})

const providerLabel = computed(() => config.value?.provider || '未设置')
const protocolLabel = computed(() => config.value?.api_protocol || 'chat/completions')
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
const canRefreshModels = computed(() => {
  return (
    canSubmit.value &&
    !modelsLoading.value &&
    Boolean(form.value.baseUrl.trim()) &&
    (Boolean(apiKeyDraft.value.trim()) || Boolean(config.value?.api_key_configured && !clearApiKey.value))
  )
})

const syncForm = (nextConfig: AdminAIConfigResponse | null) => {
  form.value = {
    provider: nextConfig?.provider || 'OpenAI Compatible',
    apiProtocol: nextConfig?.api_protocol || 'chat/completions',
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
  modelOptions.value = []

  const nextDefaults = nextConfig?.default_prompt_templates ?? {}
  const nextTemplates = nextConfig?.prompt_templates ?? {}
  for (const mode of promptModes) {
    defaultPromptTemplates.value[mode] = nextDefaults[mode] || ''
    promptTemplates.value[mode] = nextTemplates[mode] || nextDefaults[mode] || ''
  }
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
      api_protocol: form.value.apiProtocol,
      base_url: form.value.baseUrl.trim(),
      model: form.value.model.trim(),
      api_key: clearApiKey.value ? '' : apiKeyDraft.value.trim(),
      timeout: form.value.timeout.trim() || '30s',
      max_input_chars: toPositiveNumber(form.value.maxInputChars, 6000),
      max_context_chars: toNonNegativeNumber(form.value.maxContextChars, 4000),
      max_suggestions: toPositiveNumber(form.value.maxSuggestions, 3),
      enabled: form.value.enabled && !clearApiKey.value,
      clear_api_key: clearApiKey.value,
      prompt_templates: { ...promptTemplates.value }
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

const resetPromptTemplate = (mode: AdminAIPolishMode) => {
  promptTemplates.value[mode] = defaultPromptTemplates.value[mode] || ''
}

const resetAllPromptTemplates = () => {
  for (const mode of promptModes) {
    resetPromptTemplate(mode)
  }
}

const refreshModels = async () => {
  if (!canRefreshModels.value) return
  modelsLoading.value = true

  try {
    const response = await listAdminAIModels({
      provider: form.value.provider.trim(),
      api_protocol: form.value.apiProtocol,
      base_url: form.value.baseUrl.trim(),
      api_key: apiKeyDraft.value.trim() || undefined
    })
    modelOptions.value = response.data.models ?? []
    if (modelOptions.value.length && !form.value.model.trim()) {
      form.value.model = modelOptions.value[0]
    }
    toast.add({
      severity: 'success',
      summary: '模型列表已刷新',
      detail: modelOptions.value.length ? `获取到 ${modelOptions.value.length} 个模型` : '提供商没有返回模型',
      life: 2400
    })
  } finally {
    modelsLoading.value = false
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
            <span class="text-sm font-bold text-foreground">提供商名称</span>
            <AppInput v-model="form.provider" placeholder="OpenAI Compatible" :disabled="!canSubmit" />
          </label>

          <label class="block space-y-2">
            <span class="text-sm font-bold text-foreground">API 协议</span>
            <select
              v-model="form.apiProtocol"
              class="h-11 w-full rounded-full border border-border bg-surface px-4 text-sm font-semibold text-foreground transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/18 disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="!canSubmit"
            >
              <option value="chat/completions">chat/completions</option>
              <option value="responses">responses</option>
              <option value="messages">messages</option>
            </select>
          </label>

          <div class="space-y-2 md:col-span-2">
            <span class="text-sm font-bold text-foreground">Model</span>
            <div class="flex flex-col gap-2 sm:flex-row">
              <AppInput
                id="admin-ai-model"
                v-model="form.model"
                list="admin-ai-model-options"
                placeholder="gpt-4.1-mini"
                :disabled="!canSubmit"
              />
              <datalist id="admin-ai-model-options">
                <option v-for="model in modelOptions" :key="model" :value="model" />
              </datalist>
              <AppButton
                type="button"
                variant="secondary"
                class="sm:shrink-0"
                :disabled="!canRefreshModels"
                @click="refreshModels"
              >
                <RefreshCw class="size-4" aria-hidden="true" />
                {{ modelsLoading ? '获取中' : '刷新模型' }}
              </AppButton>
            </div>
          </div>

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

        <section class="mt-5 rounded-archive border border-border bg-surface-raised p-4">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="min-w-0">
              <h2 class="m-0 text-base font-black text-foreground">高级提示词</h2>
              <p class="m-0 mt-1 text-pretty text-sm text-muted-foreground">
                每个操作都可以使用独立模板，保存后随下一次润色请求生效。
              </p>
            </div>
            <AppButton
              type="button"
              variant="secondary"
              size="sm"
              class="sm:shrink-0"
              :disabled="!canSubmit"
              @click="resetAllPromptTemplates"
            >
              恢复全部默认
            </AppButton>
          </div>

          <div class="mt-4 grid gap-4">
            <label v-for="mode in promptModes" :key="mode" class="block space-y-2">
              <span class="flex items-center justify-between gap-3">
                <span class="text-sm font-bold text-foreground">{{ getAIPolishModeLabel(mode) }}</span>
                <AppButton
                  type="button"
                  variant="ghost"
                  size="sm"
                  :disabled="!canSubmit"
                  @click="resetPromptTemplate(mode)"
                >
                  恢复默认
                </AppButton>
              </span>
              <textarea
                v-model="promptTemplates[mode]"
                rows="7"
                class="w-full resize-y rounded-archive border border-border bg-surface px-4 py-3 font-mono text-xs leading-6 text-foreground outline-none transition-colors focus:border-accent focus:ring-2 focus:ring-accent/20 disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="!canSubmit"
              />
            </label>
          </div>
        </section>

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
            <dt class="text-xs font-bold text-muted-foreground">当前协议</dt>
            <dd class="m-0 mt-2 break-all text-sm font-black text-foreground">{{ protocolLabel }}</dd>
          </div>
        </dl>

        <dl class="m-0 mt-3 grid gap-3 sm:grid-cols-2">
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
