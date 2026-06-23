<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'
import { sanitizeHtml } from '@/util/sanitizeHtml'

const props = withDefaults(
  defineProps<{
    content?: string
    compact?: boolean
    preview?: boolean
  }>(),
  {
    content: '',
    compact: false,
    preview: false
  }
)

const sanitizedContent = computed(() => sanitizeHtml(props.content || ''))
const classes = computed(() =>
  cn(
    'reading-prose ck-content',
    props.compact && 'reading-prose--compact',
    props.preview && 'admin-preview-content'
  )
)
</script>

<template>
  <div :class="classes" v-html="sanitizedContent" />
</template>
