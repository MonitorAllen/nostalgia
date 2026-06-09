<script setup lang="ts">
import { computed } from 'vue'
import { AlertCircle, CheckCircle2, Info, X, AlertTriangle } from '@lucide/vue'
import { useToast, type ToastSeverity } from '@/composables/useToast'

const { toasts, remove } = useToast()

const iconMap: Record<ToastSeverity, object> = {
  success: CheckCircle2,
  info: Info,
  warning: AlertTriangle,
  error: AlertCircle,
}

const toneClass = computed(() => ({
  success: 'border-accent/30 bg-accent/10 text-accent',
  info: 'border-border bg-surface text-foreground',
  warning: 'border-warning/30 bg-warning/10 text-warning',
  error: 'border-danger/30 bg-danger/10 text-danger',
}))
</script>

<template>
  <Teleport to="body">
    <div class="fixed right-4 top-4 z-50 flex w-[min(24rem,calc(100vw-2rem))] flex-col gap-3">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="archive-glass flex items-start gap-3 rounded-archive p-4"
        :class="toneClass[toast.severity]"
        role="status"
      >
        <component :is="iconMap[toast.severity]" class="mt-0.5 h-5 w-5 shrink-0" />
        <div class="min-w-0 flex-1">
          <p class="m-0 text-sm font-bold">{{ toast.summary }}</p>
          <p v-if="toast.detail" class="m-0 mt-1 text-sm text-muted-foreground">
            {{ toast.detail }}
          </p>
        </div>
        <button
          type="button"
          class="rounded-full p-1 text-muted-foreground transition hover:bg-muted hover:text-foreground"
          aria-label="关闭提示"
          @click="remove(toast.id)"
        >
          <X class="h-4 w-4" />
        </button>
      </div>
    </div>
  </Teleport>
</template>
