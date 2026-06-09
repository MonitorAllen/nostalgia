<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    class?: string
  }>(),
  {
    class: '',
  },
)

const classes = computed(() =>
  cn('animate-pulse rounded-archive bg-muted/80', props.class),
)
</script>

<template>
  <div :class="classes" />
</template>
