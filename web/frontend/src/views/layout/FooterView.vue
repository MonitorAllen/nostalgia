<script setup>
const currentYear = new Date().getFullYear()
</script>

<template>
  <footer class="w-full py-4 bg-gray-50 mt-auto">
    <div class="flex flex-column align-items-center text-600 text-sm gap-2">
      <p class="m-0">Copyright © {{ currentYear }} Nostalgia. Design by PrimeVue.</p>
      <p class="m-0 hover:text-900 cursor-pointer transition-colors">粤ICP备2025431323号</p>
    </div>
  </footer>
</template>