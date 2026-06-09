<script setup lang="ts">
import { ChevronRight, Home } from '@lucide/vue'

defineProps<{
  items: Array<{ label: string; to?: string }>
}>()
</script>

<template>
  <nav class="archive-glass rounded-archive px-4 py-3 text-sm" aria-label="当前位置">
    <ol class="flex flex-wrap items-center gap-2 text-muted-foreground">
      <li>
        <RouterLink to="/" class="inline-flex items-center gap-1 font-semibold hover:text-accent">
          <Home class="h-4 w-4" />
          首页
        </RouterLink>
      </li>
      <li v-for="item in items" :key="item.label" class="flex items-center gap-2">
        <ChevronRight class="h-4 w-4" />
        <RouterLink v-if="item.to" :to="item.to" class="font-semibold hover:text-accent">
          {{ item.label }}
        </RouterLink>
        <span v-else class="font-bold text-foreground">{{ item.label }}</span>
      </li>
    </ol>
  </nav>
</template>
