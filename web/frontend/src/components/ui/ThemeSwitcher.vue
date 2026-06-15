<script setup lang="ts">
import { computed, type Component } from 'vue'
import { Monitor, Moon, Sun } from '@lucide/vue'
import { useTheme, type ThemeMode } from '@/composables/useTheme'

const { mode, setMode } = useTheme()

const items: Array<{ value: ThemeMode; label: string; icon: Component }> = [
  { value: 'system', label: '系统', icon: Monitor },
  { value: 'light', label: '浅色', icon: Sun },
  { value: 'dark', label: '深色', icon: Moon }
]

const activeIndex = computed(() => Math.max(0, items.findIndex((item) => item.value === mode.value)))
const themeSwitcherTranslate = computed(() => `translateX(${activeIndex.value * 2.25}rem)`)
</script>

<template>
  <div
    class="archive-glass relative inline-grid grid-cols-3 rounded-full p-1"
    role="radiogroup"
    aria-label="主题模式"
  >
    <span
      class="theme-switcher-thumb pointer-events-none absolute left-1 top-1 h-9 w-9 rounded-full bg-accent shadow-archive transition-transform duration-300 ease-out motion-reduce:transition-none"
      :style="{ transform: themeSwitcherTranslate }"
      aria-hidden="true"
    />
    <button
      v-for="item in items"
      :key="item.value"
      type="button"
      class="relative z-10 inline-flex h-9 w-9 items-center justify-center rounded-full text-xs font-bold transition-colors duration-300 ease-out motion-reduce:transition-none"
      :class="
        mode === item.value
          ? 'text-accent-foreground'
          : 'text-muted-foreground hover:bg-muted/70 hover:text-foreground'
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
