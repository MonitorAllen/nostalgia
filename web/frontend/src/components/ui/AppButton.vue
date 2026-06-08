<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
    size?: 'sm' | 'md' | 'icon'
    type?: 'button' | 'submit' | 'reset'
    disabled?: boolean
    class?: string
  }>(),
  {
    variant: 'primary',
    size: 'md',
    type: 'button',
    disabled: false,
    class: '',
  },
)

const classes = computed(() =>
  cn(
    'inline-flex items-center justify-center gap-2 rounded-full font-semibold transition-all duration-200',
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent',
    'disabled:pointer-events-none disabled:opacity-50 active:translate-y-px',
    props.size === 'sm' && 'h-9 px-3 text-sm',
    props.size === 'md' && 'h-11 px-5 text-sm',
    props.size === 'icon' && 'h-10 w-10 p-0',
    props.variant === 'primary' &&
      'bg-accent text-accent-foreground shadow-archive hover:bg-accent/90',
    props.variant === 'secondary' &&
      'border border-border bg-surface text-foreground hover:bg-muted',
    props.variant === 'ghost' &&
      'text-muted-foreground hover:bg-muted hover:text-foreground',
    props.variant === 'danger' &&
      'bg-danger text-white hover:bg-danger/90',
    props.class,
  ),
)
</script>

<template>
  <button :type="type" :disabled="disabled" :class="classes">
    <slot />
  </button>
</template>
