<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    title?: string
    icon?: object
    glass?: boolean
    class?: string
  }>(),
  {
    title: '',
    icon: undefined,
    glass: true,
    class: '',
  },
)

const classes = computed(() =>
  cn(
    'overflow-hidden rounded-archive',
    props.glass ? 'archive-glass' : 'archive-surface',
    props.class,
  ),
)
</script>

<template>
  <section :class="classes">
    <header
      v-if="title || $slots.header"
      class="flex items-center justify-between border-b border-border/70 px-4 py-3"
    >
      <slot name="header">
        <div class="flex items-center gap-2 text-sm font-bold text-foreground">
          <component :is="icon" v-if="icon" class="h-4 w-4 text-accent" />
          <span>{{ title }}</span>
        </div>
      </slot>
    </header>
    <div class="p-4">
      <slot />
    </div>
  </section>
</template>
