<script setup lang="ts">
const currentYear = new Date().getFullYear()
</script>

<template>
  <footer class="border-t border-border/70 px-4 py-8">
    <div class="mx-auto flex max-w-7xl flex-col gap-2 text-sm text-muted-foreground md:flex-row md:items-center md:justify-between">
      <p class="m-0 font-semibold">Copyright © {{ currentYear }} Nostalgia.</p>
      <p class="m-0">粤ICP备2025431323号 · archived with care</p>
    </div>
  </footer>
</template>
