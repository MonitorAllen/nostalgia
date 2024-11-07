<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import Toast from 'primevue/toast'
import NavBar from '@/views/layout/NavBar.vue'

const route = useRoute()
</script>

<template>
  <div class="card flex justify-center">
    <Toast />
  </div>
  <header>
    <div class="top">
      <NavBar v-if="!route.meta.hideNavbar"></NavBar>
    </div>
  </header>
  <RouterView />
</template>

<style scoped></style>
