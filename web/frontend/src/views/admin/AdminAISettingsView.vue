<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Bot, CheckCircle2, KeyRound, Settings2, ShieldCheck, XCircle } from '@lucide/vue'
import { getAdminAIConfig } from '@/admin/api/adminAiApi'
import type { AdminAIConfigResponse } from '@/admin/types'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'

const router = useRouter()
const config = ref<AdminAIConfigResponse | null>(null)
const loading = ref(false)

const providerLabel = computed(() => config.value?.provider || '未设置')
const modelLabel = computed(() => config.value?.model || '未设置')
const baseUrlLabel = computed(() => config.value?.base_url || '未设置')
const sourceLabel = computed(() =>
  config.value?.source === 'runtime_env' ? '运行时环境变量' : config.value?.source || '未设置'
)

const loadConfig = async () => {
  if (loading.value) return
  loading.value = true

  try {
    const response = await getAdminAIConfig()
    config.value = response.data
  } catch {
    config.value = null
  } finally {
    loading.value = false
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
          <Settings2 class="size-4" aria-hidden="true" />
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
      <div class="archive-surface rounded-archive p-5">
        <div class="grid gap-3 md:grid-cols-2">
          <div class="rounded-archive border border-border bg-surface-raised p-4">
            <p class="m-0 text-xs font-bold text-muted-foreground">Provider</p>
            <p class="m-0 mt-2 break-all text-sm font-black text-foreground">{{ providerLabel }}</p>
          </div>
          <div class="rounded-archive border border-border bg-surface-raised p-4">
            <p class="m-0 text-xs font-bold text-muted-foreground">Model</p>
            <p class="m-0 mt-2 break-all text-sm font-black text-foreground">{{ modelLabel }}</p>
          </div>
          <div class="rounded-archive border border-border bg-surface-raised p-4 md:col-span-2">
            <p class="m-0 text-xs font-bold text-muted-foreground">Base URL</p>
            <p class="m-0 mt-2 break-all text-sm font-black text-foreground">{{ baseUrlLabel }}</p>
          </div>
        </div>

        <dl class="m-0 mt-5 grid gap-3 sm:grid-cols-3">
          <div class="rounded-archive border border-border p-4">
            <dt class="text-xs font-bold text-muted-foreground">请求超时</dt>
            <dd class="m-0 mt-2 text-lg font-black text-foreground">{{ config?.timeout || '未设置' }}</dd>
          </div>
          <div class="rounded-archive border border-border p-4">
            <dt class="text-xs font-bold text-muted-foreground">输入上限</dt>
            <dd class="m-0 mt-2 text-lg font-black text-foreground tabular-nums">
              {{ config?.max_input_chars || 0 }}
            </dd>
          </div>
          <div class="rounded-archive border border-border p-4">
            <dt class="text-xs font-bold text-muted-foreground">候选数量</dt>
            <dd class="m-0 mt-2 text-lg font-black text-foreground tabular-nums">
              {{ config?.max_suggestions || 0 }}
            </dd>
          </div>
        </dl>
      </div>

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
