<script setup lang="ts">
import {computed} from "vue";

const props = defineProps<{
  content: string // 需要对关键词高亮的内容
  keyword?: string // 关键词
  highlightClass?: string // 自定义高亮样式类名
}>()

// 正则特殊字符转义（防止用户输入 * ？+ 等导致正则崩溃）
const escapeRegExp = (s: string): string => {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

const highlightedHtml = computed(() => {
  if (!props.keyword || !props.keyword.trim()) {
    return props.content;
  }

  // 按空格拆分关键词，并过滤空格,最后进行转义
  const words = props.keyword.split(/\s+/).filter(w => w.length > 0).map(w => escapeRegExp(w))
  if (words.length === 0) {
    return props.content
  }

  // 构建正则，全局匹配 + 忽略大小写
  const regex = new RegExp(`(${words.join('|')})`, 'gi')

  // 替换文本
  // $1 代表正则捕获到的原文（保留原文的大小写），包裹在高亮标签中
  const className = props.highlightClass || 'text-green-500 font-bold bg-green-50 border-round-md px-1'
  return props.content.replace(regex, `<span class="${className}">$1</span>`)
})

</script>

<template>
  <span v-html="highlightedHtml"></span>
</template>

<style scoped>

</style>