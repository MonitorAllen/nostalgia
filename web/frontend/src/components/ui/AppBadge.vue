<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    tone?: 'neutral' | 'accent' | 'danger' | 'warning'
    class?: string
  }>(),
  {
    tone: 'neutral',
    class: '',
  },
)

const classes = computed(() =>
  cn(
    'inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs font-semibold leading-none',
    props.tone === 'neutral' && 'border-border bg-surface-raised text-muted-foreground',
    props.tone === 'accent' && 'border-accent/30 bg-accent/10 text-accent',
    props.tone === 'danger' && 'border-danger/30 bg-danger/10 text-danger',
    props.tone === 'warning' && 'border-warning/30 bg-warning/10 text-warning',
    props.class,
  ),
)
</script>

<template>
  <span :class="classes">
    <slot />
  </span>
</template>
