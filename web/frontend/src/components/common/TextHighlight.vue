<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  content: string // 需要对关键词高亮的内容
  keyword?: string // 关键词
  highlightClass?: string // 自定义高亮样式类名
}>()

const escapeRegExp = (s: string): string => {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

const escapeHtml = (s: string): string => {
  const entities: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  }

  return s.replace(/[&<>"']/g, (char) => entities[char])
}

const highlightedHtml = computed(() => {
  if (!props.keyword || !props.keyword.trim()) {
    return escapeHtml(props.content)
  }

  const words = props.keyword.split(/\s+/).filter((w) => w.length > 0).map((w) => escapeRegExp(w))
  if (words.length === 0) {
    return escapeHtml(props.content)
  }

  const regex = new RegExp(`(${words.join('|')})`, 'gi')
  const className = escapeHtml(props.highlightClass || 'rounded-md bg-accent/10 px-1 font-bold text-accent')

  return props.content
    .split(regex)
    .map((part) => {
      if (part.match(regex)) {
        return `<span class="${className}">${escapeHtml(part)}</span>`
      }
      return escapeHtml(part)
    })
    .join('')
})

</script>

<template>
  <span v-html="highlightedHtml"></span>
</template>

<style scoped>

</style>
