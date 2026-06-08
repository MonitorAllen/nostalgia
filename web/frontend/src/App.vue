<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import NavBar from '@/views/layout/NavBar.vue'
import FooterView from './views/layout/FooterView.vue'
import ToastViewport from '@/components/ui/ToastViewport.vue'

const route = useRoute()
</script>

<template>
  <ToastViewport />
  <div class="flex min-h-screen w-full flex-col">
    <header v-if="!route.meta.hideNavbar" class="sticky top-0 z-40">
      <NavBar />
    </header>
    <main class="w-full flex-1">
      <RouterView />
    </main>
    <footer>
      <FooterView />
    </footer>
  </div>
</template>
