<script setup lang="ts">
import { Monitor, Moon, Sun } from '@lucide/vue'
import { useTheme, type ThemeMode } from '@/composables/useTheme'

const { mode, setMode } = useTheme()

const items: Array<{ value: ThemeMode; label: string; icon: object }> = [
  { value: 'system', label: '系统', icon: Monitor },
  { value: 'light', label: '浅色', icon: Sun },
  { value: 'dark', label: '深色', icon: Moon }
]
</script>

<template>
  <div class="archive-glass inline-flex rounded-full p-1" role="radiogroup" aria-label="主题模式">
    <button
      v-for="item in items"
      :key="item.value"
      type="button"
      class="inline-flex h-9 w-9 items-center justify-center rounded-full text-xs font-bold transition"
      :class="
        mode === item.value
          ? 'bg-accent text-accent-foreground'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
      "
      role="radio"
      :aria-checked="mode === item.value"
      :aria-label="`${item.label}主题`"
      :title="`${item.label}主题`"
      @click="setMode(item.value)"
    >
      <component :is="item.icon" class="h-4 w-4" />
    </button>
  </div>
</template>
